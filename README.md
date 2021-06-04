# rssd
#### Poll and execute command when there is a new rss item  
*There is still much to do, keep your eyes open*

## Installation
```
make
make install
```

## Usage
### Config
Default location is `$XDG_CONFIG_HOME/rssd/config.json` but could be overridden with `--config`
```
rssd --config ./cfg/cfg.json
```

### Adding a feed
`rssd add-feed "https://ssiyad.com/blog/index.xml"`

### Listing current feeds
`rssd list-feed`
and output will be something like,
```
+-------+--------------------------------+-----------------------------------+
| INDEX |              FEED              |               LAST                |
+-------+--------------------------------+-----------------------------------+
|     0 | http://rss.art19.com/the-daily | https://www.nytimes.com/the-daily |
+-------+--------------------------------+-----------------------------------+
```

### Removing a feed
You could remove a feed using it's index  
`rssd remove-feed 0`

### Setting the command to execute
`rssd set-exec command`

#### Examples
- Desktop notifications  
`rssd set-exec "notify-send '&title' '&item_title'"`
- Telegram bot  
`rssd set-exec "https://api.telegram.org/bot\$BOT_TOKEN/sendMessage?chat_id=\$TG_CHAT&text=&item_title"`

### Running rssd
A systemd timer and unit is provided or using `--standalone`
```
rssd --standalone --interval 1
```
`--interval` is in minutes and used only in standalone mode

## Available placeholders
Placeholders need to be prefixed with `&`, like `&item_title`
```
title
desc
lang
item_title
item_link
item_pubDate
item_desc
item_authorName
item_authorEmail
```

## TODO
- additional flags
- additional examples