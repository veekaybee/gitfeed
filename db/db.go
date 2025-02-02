package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

type ATPost struct {
	Did    string `json:"did"`
	TimeUs int64  `json:"time_us"`
	Type   string `json:"type"`
	Kind   string `json:"kind"`
	Commit struct {
		Rev        string `json:"rev"`
		Type       string `json:"type"`
		Operation  string `json:"operation"`
		Collection string `json:"collection"`
		Rkey       string `json:"rkey"`
		Record     struct {
			Type      string    `json:"$type"`
			CreatedAt time.Time `json:"createdAt"`
			Embed     struct {
				Type     string `json:"$type"`
				External struct {
					Description string `json:"description"`
					Title       string `json:"title"`
					URI         string `json:"uri"`
				} `json:"external"`
			} `json:"embed"`
			Facets []struct {
				Features []struct {
					Type string `json:"$type", omitempty`
					URI  string `json:"uri", omitempty`
				} `json:"features"`
				Index struct {
					ByteEnd   int `json:"byteEnd"`
					ByteStart int `json:"byteStart"`
				} `json:"index"`
			} `json:"facets"`
			Langs []string `json:"langs", omitempty`
			Text  string   `json:"text"`
		} `json:"record"`
		Cid string `json:"cid"`
	} `json:"commit"`
}

type DBPost struct {
	Did        string
	TimeUs     int64
	Kind       string
	Rev        string
	Operation  string
	Collection string
	Rkey       string
	Type       string
	CreatedAt  time.Time
	Langs      sql.Null[string]
	ParentCid  string
	ParentURI  string
	RootCid    string
	RootURI    string
	Text       string
	Cid        string
	ID         string
	URI        string
}

func InitDB() (*sql.DB, error) {

	var err error
	var gitfeed = "gitfeed.db"

	DB, err := sql.Open("sqlite3", gitfeed)
	if err != nil {
		panic(err)
	}

	_, err = DB.Exec(`PRAGMA journal_mode=WAL;`)
	if err != nil {
		fmt.Println("Error setting WAL mode:", err)
		panic(err)
	}

	_, err = DB.Exec(`PRAGMA busy_timeout = 5000;`)
	if err != nil {
		fmt.Println("Error setting WAL mode:", err)
		panic(err)
	}

	err = DB.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging database: %v", err)
	}

	fmt.Println("Connected to database:", gitfeed)
	return DB, nil
}

func (pr *PostRepository) CreateTableIfNotExists(tableName string, columns map[string]string) error {

	var columnDefs []string
	for colName, colType := range columns {
		columnDefs = append(columnDefs, fmt.Sprintf("%s %s", colName, colType))
	}

	fmt.Printf("Creating table %s", tableName)

	// Construct the CREATE TABLE query
	query := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            %s
        )`,
		tableName,
		strings.Join(columnDefs, ",\n"))
	fmt.Println(query)

	// Create table
	_, err := pr.db.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating table %s: %v", tableName, err)
	}

	// Index table on timestamp
	_, err = pr.db.Exec("CREATE INDEX IF NOT EXISTS time_us ON posts(time_us);")
	if err != nil {
		return fmt.Errorf("error creating index time_us: %v", err)
	}

	fmt.Printf("Table '%s' created or already exists\n", tableName)
	return nil
}

type PostRepository struct {
	db   *sql.DB
	lock *sync.Mutex
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db, lock: &sync.Mutex{}}
}

type PostRepo interface {
	GetPost(uuid string) (*DBPost, error)
	WritePost(p DBPost) error
	DeletePost(uuid string) error
	DeletePosts() error
	GetAllPosts() ([]DBPost, error)
	GetTimeStamp() (int64, error)
}

func (pr *PostRepository) GetPost(did string) (*DBPost, error) {
	pr.lock.Lock()
	defer pr.lock.Unlock()

	sqlStmt := `SELECT *
                FROM posts 
                WHERE did = $1`

	var post DBPost
	err := pr.db.QueryRow(sqlStmt, did).Scan(
		&post.Did,
		&post.TimeUs,
		&post.Kind,
		&post.Rev,
		&post.Operation,
		&post.Collection,
		&post.Rkey,
		&post.Cid,
		&post.Type,
		&post.CreatedAt,
		&post.Langs,
		&post.Text,
		&post.ID,
		&post.URI,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no post found with DID: %s", did)
		}
		return nil, fmt.Errorf("error querying post: %w", err)
	}

	return &post, nil
}

func (pr *PostRepository) WritePost(p DBPost) error {
	pr.lock.Lock()
	defer pr.lock.Unlock()
	sqlStmt := `INSERT INTO posts (did, 
	time_us, 
	kind, 
	commit_rev, 
	commit_operation, 
	commit_collection, 
	commit_rkey,
	commit_cid, 
	record_type, 
	record_created_at, 
	record_langs, 
	record_text,
	record_uri)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,$13)`

	_, err := pr.db.Exec(sqlStmt,
		p.Did,
		p.TimeUs,
		p.Kind,
		p.Rev,
		p.Operation,
		p.Collection,
		p.Rkey,
		p.Cid,
		p.Type,
		p.CreatedAt,
		p.Langs,
		p.Text,
		p.URI)
	if err != nil {
		log.Printf("%+v\n", p)
		return fmt.Errorf("could not write to db: %w", err)
	}
	log.Printf("wrote %s\n", p.Did)
	return nil
}

func (pr *PostRepository) DeletePost(uuid string) error {
	pr.lock.Lock()
	defer pr.lock.Unlock()
	sqlStmt := `DELETE FROM posts WHERE postid = $1`

	_, err := pr.db.Exec(sqlStmt, uuid)
	if err != nil {
		return fmt.Errorf("could not delete from db: %w", err)
	}

	return nil
}

func (pr *PostRepository) DeletePosts() error {
	pr.lock.Lock()
	defer pr.lock.Unlock()

	log.Printf("Deleting old posts...")
	sqlStmt := `DELETE FROM posts WHERE NOT EXISTS (
    SELECT 1 FROM (
        SELECT * FROM posts ORDER BY time_us DESC LIMIT 10
    ) AS temp WHERE posts.did = temp.did AND posts.time_us = temp.time_us
	);`

	_, err := pr.db.Exec(sqlStmt)
	if err != nil {
		return fmt.Errorf("could not delete from db: %w", err)
	}

	return nil
}

func (pr *PostRepository) GetAllPosts() ([]DBPost, error) {
	pr.lock.Lock()
	defer pr.lock.Unlock()

	log.Printf("Fetching top 10 posts desc from DB...")
	sqlStmt := `SELECT  DISTINCT did, 
	                             time_us, 
								 kind, 
								 commit_rev, 
								 commit_operation, 
								 commit_collection, 
                                 commit_rkey, 
								 record_type, 
								 record_created_at, 
								 record_langs, 
								 commit_cid, 
								 record_text, 
								 record_uri  
								 FROM posts
				                 ORDER BY time_us desc LIMIT 10;`

	rows, err := pr.db.Query(sqlStmt)
	if err != nil {
		return nil, fmt.Errorf("error querying posts: %w", err)
	}

	var posts []DBPost

	log.Printf("Iterating on rows...")
	for rows.Next() {
		var p DBPost

		err := rows.Scan(
			&p.Did,
			&p.TimeUs,
			&p.Kind,
			&p.Rev,
			&p.Operation,
			&p.Collection,
			&p.Rkey,
			&p.Type,
			&p.CreatedAt,
			&p.Langs,
			&p.ParentCid,
			&p.Text,
			&p.URI,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning post: %w", err)
		}
		posts = append(posts, p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating posts: %w", err)
	}

	if len(posts) == 0 {
		return nil, fmt.Errorf("no posts found")
	}

	return posts, nil

}

func (pr *PostRepository) GetTimeStamp() (int64, error) {
	pr.lock.Lock()
	defer pr.lock.Unlock()
	sqlStmt := `SELECT time_us FROM posts ORDER BY time_us DESC LIMIT 1;`
	var timeUs int64
	if err := pr.db.QueryRow(sqlStmt).Scan(&timeUs); err != nil {

		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no posts found")
		}
		return 0, err
	}

	return timeUs, nil
}
