# rssd
#### Poll and execute command when there is a new rss item  
*There is still much to do, keep your eyes open*

## Building
`go build`

## Usage
### Adding a feed
`rssd add-feed "https://ssiyad.com/blog/index.xml"`

### Setting the command to execute
`rssd set-exec command`

#### Examples
- Desktop notifications  
`rssd set-exec "notify-send '&title' '&item_title'"`
- Telegram bot  
`rssd set-exec "https://api.telegram.org/bot\$BOT_TOKEN/sendMessage?chat_id=\$TG_CHAT&text=&item_title"`

## Available placeholders
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
- systemd unit
- systemd timer
- list feed
- remove feed
- additional flags
- additional examples