# simplewebwatcher

## Installation
```bash
go get github.com/fekle/simplewebwatcher
``

## Update
```bash
go get -u github.com/fekle/simplewebwatcher
```

## Usage
Running `simplewebwatcher` for the first time creates the neccessary folder (`~/.simplewebwatcher`) and writes a sample configuration to `~/.simplewebwatcher/config`
Currently, `simplewebwatcher` only has one mode, called `cron`, for use with cronjobs and other schedulers.

## launchd (OSX)
Edit `scripts/simplewebwatcher.plist`, copy the file to `~/Library/LaunchAgents/simplewebwatcher.plist`.
Run `launchctl load ~/Library/LaunchAgents/simplewebwatcher.plist` to load the service, and `launchctl unload ~/Library/LaunchAgents/simplewebwatcher.plist` to unload it.
