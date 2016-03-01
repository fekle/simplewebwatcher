# simplewebwatcher

## Installation
```bash
go get github.com/fekle/simplewebwatcher
```

## Update
```bash
go get -u github.com/fekle/simplewebwatcher
```

## Usage
Running `simplewebwatcher` for the first time creates the neccessary folder (`~/.simplewebwatcher`) and writes a sample configuration to `~/.simplewebwatcher/config`.

Currently, `simplewebwatcher` only has one mode, which fetches all configured sites, compares their sizes and hashes to the data stored in the configuration. On a mismatch, simplewebwatcher logs the occurence and alerts the user, before updating the configuration with the new data.

As the nameimplies, this mode is designed for use with cronjobs and other schedulers.

## launchd (OSX)
Edit `scripts/simplewebwatcher.plist`, copy the file to `~/Library/LaunchAgents/simplewebwatcher.plist`.

Run `launchctl load ~/Library/LaunchAgents/simplewebwatcher.plist` to load the service, and `launchctl unload ~/Library/LaunchAgents/simplewebwatcher.plist` to unload it.

## Configuration Sample

```toml
# Site Block, can be repeated for configuring multiple pages
[[Site]]
  Description = "First" # Description used in popup and log
  URL = "http://localhost" # Url scheme:   http[s]://example.com[:123]
  Username = "user" # leave blank to disable http auth
  Password = "password" # leave blank to disable http auth
  LastCheck = 2016-03-01T19:48:40Z # don't edit - used to store time of last check
  LastBytes = 0 # don't edit - used to store size of last check
  LastHash = "" # don't edit - used to store hash of last check

[[Site]]
  Description = "Second"
  URL = "http://localhost"
  Username = "user"
  Password = "password"
  LastCheck = 2016-03-01T19:48:40Z
  LastBytes = 0
  LastHash = ""

[[Site]]
  Description = "Third"
  URL = "http://localhost"
  Username = "user"
  Password = "password"
  LastCheck = 2016-03-01T19:48:40Z
  LastBytes = 0
  LastHash = ""

```
