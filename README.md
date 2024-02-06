## Description:

This Docker image is designed for use with Drone CI as a plugin to send messages to a Nextcloud Talk room that has a bot.

## Usage:

To use this Docker image in your Drone CI pipeline, specify it as a step in your .drone.yml file. Configure the necessary environment variables to customize the message and target room.

Required environment variables are:

- nextcloud_server_url: URL of your Nextcloud instance
- bot_secret: Bot secret you gave on the creation of the bot
- room_id: You can get this ID from the URL: `http://nextcloud.example/call/{roomID}`

Message variable is not required and the default looks like this:

```
Status: ❌ or ✅ and **DRONE_BUILD_STATUS**
Branch: [DRONE_BRANCH](DRONE_REPO_LINK)
Commit: DRONE_COMMIT
Author: DRONE_COMMIT_AUTHOR
Hash: [DRONE_COMMIT_SHA](DRONE_COMMIT_LINK)
[View full log here](DRONE_BUILD_LINK)
```

### Example:

```
steps:
- name: notify-nextcloud-talk
  image: malina01/notification-to-talk
  settings:
    nextcloud_server_url: http://nextcloud.example.com
    bot_secret: your_bot_secret
    message: Build completed successfully
    room_id: your_room_id

```

### Dockerfile

To create image you first need to create bin version of file with command:

```
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o webhook
```
