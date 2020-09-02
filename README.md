# techmeetup
A CLI tool that provides various utility capabilities for automating aspects of tech meetup groups

This is another attempt to automate bit and pieces of workflow while managing a tech meetup group. Unfortunately, pass efforts are done up in python, making it slightly difficult to distribute. At the same time, due to the imperative nature of the scripts - scripts only create resources on destination platforms without checking; onus on the person who is running the script. This made it quite difficult to run it

As part of revamping it, this cli tool is taking a page from how Kubernetes does things. Have settings be declared and the binary will somehow make it happen. (We're not building it on k8s - there is no need to have scale etc for this)

# Building the CLI

```bash
go build -o techmeetup ./cmd
```

# Planned Features

- GDG Cloud Singapore Utility
  - bitly link creation - connect to Googleslides and retrieve all http link looking things and grab it to a yaml/json file
  - Allow user to replace it in one swoop (User to pass in a yaml/json file that would alter the links accordingly)
- GDG Cloud Singapore meetup management
  - Handle calendar invites
    - Read calendar invites for event
    - Write calendar invites for events
  - Update googlesheets (To have a update of sorts to other members)
    - Read events from googlesheets
    - Write events to googlesheets
  - To update meetup.com
    - Read events from meetup.com
    - Write events into meetup.com
  - To update streamyard
    - Read events from streamyard
    - Write/update events into streamyard
  - To update website
    - Read events from meetup.com
    - Write events into github.com
  - To facebook groups
    - Read events from facebook groups
    - Write events into facebook group
  - To facebook page
    - Read event from facebook page
    - Write events into facebook page
  - To Slack channel
    - Read chat from Slack group
    - Write event into Slack group
  - To linkedin
    - Read posts from page
    - Write posts into page
- Sync assets between the following sources/destinations
  - Google Photo Album
  - Facebook Group Album
  - Facebook Page Album
  - Meetup.com

