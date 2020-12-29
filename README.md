# TiddlyWiki Server

Note: Original written by Russ Cox, I simply edited it to save to local
filesystem.

This is a app, written in Go, that can serve
as the back end for a personal [TiddlyWiki](http://tiddlywiki.com/)

The [TiddlyWiki5](https://github.com/Jermolene/TiddlyWiki5) implementation
has a number of back end options. This app implements the backend expected
by the “[TiddlyWeb](http://tiddlyweb.com/) and TiddlySpace components” plugin.

## Authentication

No Authentication, it assumes that all users have full access to everything

## Data model

Every "thing" gets stored as a file on the disk where you run this program.

## Deployment

Put this program where you want all the files to be saved, and run with

	./tiddly -localhost -port 8080

Then visit http://localhost:8080/.

If you deploy without -localhost, anyone with access to the url can do anything.

## Plugins

TiddlyWiki supports extension through plugins. 
Plugins need to be in the downloaded index.html, not lazily
like other tiddlers. Therefore, installing a plugin means 
updating a local copy of index.html and redeploying it to
the server.

## Macros

TiddlyWiki allows tiddlers with the tag `$:/tags/Macro` to contain
global macro definitions made available to all tiddlers.
The lazy loading of tiddler bodies interferes with this: something
has to load the tiddler body before the macros it contains take effect.
To work around this, the app includes the body of all macro tiddlers
in the initial tiddler list (which otherwise does not contain bodies).
This is sufficient to make macros take effect on reload.

For some reason, no such special hack is needed for `$:/tags/Stylesheet` tiddlers.

## Synchronization

If you set Control Panel > Info > Basics > Default tiddlers by clicking
“retain story ordering”, then the story list (the list of tiddlers shown on the page)
is written to the server as it changes and is polled back from the server every 60 seconds.
This means that if you have the web site open in two different browsers 
(for example, at home and at work), changes to what you're viewing in one
propagate to the other.

## TiddlyWiki base image

The TiddlyWiki code is stored in and served from index.html, which
(as you can see by clicking on the Tools tab) is TiddlyWiki version 5.1.21.

Plugins must be pre-baked into the TiddlyWiki file, not stored on the server
as lazily loaded Tiddlers. The index.html in this directory is 5.1.21 with
the TiddlyWeb and Markdown plugins added. The TiddlyWeb plugin is
required, so that index.html talks back to the server for content.

The process for preparing a new index.html is:

- Open tiddlywiki-5.1.21.html in your web browser.
- Click the control panel (gear) icon.
- Click the Plugins tab.
- Click "Get more plugins".
- Click "Open plugin library".
- Type "tiddlyweb" into the search box. The "TiddlyWeb and TiddlySpace components" should appear.
- Click Install. A bar at the top of the page should say "Please save and reload for the changes to take effect."
- Click the icon next to save, and an updated file will be downloaded.
- Open the downloaded file in the web browser.
- Repeat, adding any more plugins.
- Copy the final download to index.html.

