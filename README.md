# techmeetup

A CLI tool that provides various utility capabilities for automating aspects of tech meetup groups

This is another attempt to automate bit and pieces of workflow while managing a tech meetup group. Unfortunately, pass efforts are done up in python, making it slightly difficult to distribute. At the same time, due to the imperative nature of the scripts - scripts only create resources on destination platforms without checking; onus on the person who is running the script. This made it quite difficult to run it

As part of revamping it, this cli tool is taking a page from how Kubernetes does things. Have settings be declared and the binary will somehow make it happen. (We're not building it on k8s - there is no need to have scale etc for this)

# Usage of tool

There would little to no documentation for this tool. Most of the documentation will be part of the cli - this would make it easier to understand what it is probably doing as well

# Building the CLI

```bash
go build -o techmeetup ./cmd
```

# Completed Features

- GDG Cloud Singapore Utility
  - bitly link creation - connect to Googleslides and retrieve all http link looking things and grab it to a yaml/json file
  - Allow user to replace it in one swoop (User to pass in a yaml/json file that would alter the links accordingly)

# In progress features

- Handle calendar invites -> need to consider speakers/organizers
  - Read calendar invites for event
  - Create calendar invites for events
  - Update calendar invites for events
- To update streamyard
  - Read events from streamyard - Done, require integration
  - Create event in streamyard - Done but need to consider facebook destination
  - Update event in streamyard
- To update meetup.com
  - Read events from meetup.com - Done, require integration
  - Create events into meetup.com
  - Update events into meetup.com

# Planned Features

- GDG Cloud Singapore meetup management

  - NOTE: For all the below mentioned features:
    - Features should have an end sync date of sorts (Make sure that slides don't update an most critical moment)
    - Allow user to hit an endpoint to force update right now
  - Backup of settings
  - Create the biweekly meetup slides
    - Read and maintain state of what has already been created
    - Alter the slides accordingly based on new information
  - Update googlesheets (To have a update of sorts to other members)
    - Read events from googlesheets
    - Write events to googlesheets
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
