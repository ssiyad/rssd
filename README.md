# rssd
#### Poll and execute command when there is a new rss item  
*There is still much to do, keep your eyes open*

## Installation
```
> make
> make install
```

## Usage
### Config
Default location is `$XDG_CONFIG_HOME/rssd/config.json` but could be overridden with `--config`
```
> rssd --config ./cfg/cfg.json
```

### Adding a feed
```
> rssd add-feed "https://ssiyad.com/blog/index.xml"
```

### Listing current feeds
```
> rssd list-feed
+-------+--------------------------------+-----------------------------------+
| INDEX |              FEED              |               LAST                |
+-------+--------------------------------+-----------------------------------+
|     0 | http://rss.art19.com/the-daily | https://www.nytimes.com/the-daily |
+-------+--------------------------------+-----------------------------------+
```

### Removing a feed
```
> rssd remove-feed 0
removed:  https://www.twentyfournews.com/feed
```

### Setting the command to execute
```
> rssd set-exec command
```

#### Examples
- Desktop notifications  
    ```
    > rssd set-exec "notify-send '&title' '&item_title'"
    ```
- Telegram bot  
    ```
    > rssd set-exec "https://api.telegram.org/bot\$BOT_TOKEN/sendMessage?chat_id=\$TG_CHAT&text=&item_title"
    ```

### Running rssd
- using systemd  
    ```
    > systemctl --user enable rssd.timer
    > systemctl --user start rssd.timer
    ```
    *`make install` copy service and unit files into `/usr/lib/systemd/user`*
- with standalone mode
    ```
    > rssd --standalone --interval 5
    ```
    *`--interval` is in minutes and is used only when in standalone mode*

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
