# scw-bot

This is a Discord bot hosted on asdasd.hu. It's handling start/stop of Scaleway instances/services (only VPN at the moment). This relies on a set of Shell scripts that do the actual SCW integration. TODO: publish those as well or rewrite to Go and add here.

## Usage

You need to mention the bot with `@scw-bot`.

## Deployment

### First deployment

Install the `scw-bot.service` systemd service. Need to edit to add your Discord token!

### Redeploy

```sh
go build && \
    ssh asdasd.hu systemctl stop scw-bot && \
    scp scw-bot asdasd.hu:/root/scw-bot && \
    ssh asdasd.hu systemctl start scw-bot
```