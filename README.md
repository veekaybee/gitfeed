<img src="static/android-chrome-512x512.png" alt="drawing" width="400"/>

# GitFeed

A Go-based web app that collects all posts with GitHub links from the [Bluesky Jetstream](https://docs.bsky.app/blog/jetstream) by chronological order. It displays the last 10 links in a reverse-chronological stream so you can see what GitHub repos people are chatting about on Bluesky. There are about ~100 of these events per day, collected via websocket. 

> [!NOTE]  
> This app is experimental. Expect breaking API changes up until minor semver release. ⚡️ 

<div style="text-align: center;"><img src="/ui.png" alt="drawing" width="400"/></div>

## Running: 

Gitfeed consists of a post ingest/delete process and a frontend. 

1. `git clone`
2.  Kill all instances of running application
3.  Run application 
    + `make run-serve`  # Runs the API and frontend
    + `make run-ingest` # Runs the ingest from the jetstream

## Developing:

Gitfeed inclues a Go API that abstracts the repository pattern over a SQLite db. Code can be built and deployed using Go binaries. 
The frontend is pure CSS/HTML/Javascript and can be developed ad-hoc without npm. 

<div style="text-align: center;"><img src="/architecture.png" alt="drawing" width="400"/></div>

To run both, build the Go binary to ingest the data and start the front-end to serve the app and the API. 

1. `git clone`
2. Backend lives in `cmd`, 'routes', `handlers`, and `db`
3. Frontend lives in `static`
4. Use makefile-based development: `make run-serve`, `make run-ingest` for quick dev loops.


## About the Jetstream and data model 

Bluesky is an open social network running on [the AT Protocol](https://github.com/bluesky-social/pds?tab=readme-ov-file#what-is-at-protocol) that [operates a firehose](https://docs.bsky.app/docs/advanced-guides/firehose), an authenticated stream of events with all activity on the network. Anyone can consume/produce against the firehose [following the developer guidelines.](https://docs.bsky.app/docs/support/developer-guidelines). 

One of the key features developers can do is [create custom feeds of content](https://docs.bsky.app/docs/starter-templates/custom-feeds) based on either simple heuristics like regex, or collecting data from the firehose for machine-learning style feeds including lookup with embedding models, activity aggregation, etc.  

This repo started exploring the idea of creating a custom feed and publishing it to my own PDS [in Go](https://github.com/veekaybee/gitfeed/blob/main/publishXRPC.go). It since moved to consuming directly from Jetstream, a lighter (and less stable) implementation that doesn't include the full scope of history that the firehose does [for every entry in a user's PDS.](https://jazco.dev/2024/09/24/jetstream/.) The tradeoff is that we are now consuming untrusted elements; however, since this is from the bluesky social relay, it's fine in ways that it might not be for other, less official relays.   

## Looking at events in the Jetstream

You can check GitHub events streaming to jetstream with: 

```sh
websocat wss://jetstream2.us-east.bsky.network/subscribe\?wantedCollections=app.bsky.feed.post | grep "github" | jq .
```

## Data Model:

AtProto has its own data model, defined using [schemas called "Lexicons"](https://atproto.com/guides/lexicon). For posts and actions, they look like this. 

```json5
{
  "did": "did:plc:eygmaihciaxprqvxpfvl6flk",
  "time_us": 1725911162329308,
  "kind": "commit",
  "commit": {
    "rev": "3l3qo2vutsw2b",
    "operation": "create",
    "collection": "app.bsky.feed.like",
    "rkey": "3l3qo2vuowo2b",
    "record": {
      "$type": "app.bsky.feed.like",
      "createdAt": "2024-09-09T19:46:02.102Z",
      "subject": {
        "cid": "bafyreidc6sydkkbchcyg62v77wbhzvb2mvytlmsychqgwf2xojjtirmzj4",
        "uri": "at://did:plc:wa7b35aakoll7hugkrjtf3xf/app.bsky.feed.post/3l3pte3p2e325"
      }
    },
    "cid": "bafyreidwaivazkwu67xztlmuobx35hs2lnfh3kolmgfmucldvhd3sgzcqi"
  }
}
```

DID is the ID of the PDS (user repository) where the action happened, the record type of "app.bsky.feed.post" is what we care about, and each record has both a text entry, which truncates the text, [and a facet](https://docs.bsky.app/docs/advanced-guides/post-richtext), which has all the contained links and rich text elements in the post. 

Here's a full example of a GitHub link post:

```sh
{
  "did": "did:plc:",
  "time_us": 1732988544395778,
  "type": "com",
  "kind": "commit",
  "commit": {
    "rev": "",
    "type": "c",
    "operation": "create",
    "collection": "app.bsky.feed.post",
    "rkey": "",
    "record": {
      "$type": "app.bsky.feed.post",
      "createdAt": "2024-11-29T17:42:14.541Z",
      "embed": {
        "$type": "app.bsky.embed.external",
        "external": {
          "description": "",
          "thumb": {
            "$type": "blob",
            "ref": {
              "$link": ""
            },
            "mimeType": "image/jpeg",
            "size": 
          },
          "title": "",
          "uri": ""
        }
      },
      "facets": [
        {
          "features": [
            {
              "$type": "app.bsky.richtext.facet#link",
              "uri": ""
            }
          ],
          "index": {
            "byteEnd": 85,
            "byteStart": 54
          }
        }
      ],
      "langs": [
        "en"
      ],
      "text": "..."
    },
    "cid": ""
  }
}
```

## Thanks

Thanks to: AtProto devs, the [Bluesky docs](https://docs.bsky.app/) and everyone in the [API Touchers Discord](discord.gg/FS9U8A7F) who helped by patiently answering questions about the data model. 
