### Changes required
- [ ] Make it more usable.
    - More links in a screen
    - do not need the box and whitespaces to the sides
    - the checkbox seems unecessary
    - search is not working properly
    - duplicates appear on a selection
- [ ] search seems to be working incorrectly
- [ ] after you reset the search the index doesnt go back to old contents
- [ ] put the data into an instore database with backup instead of files
- [ ] change the backend datastore to be postgres intead of local files
    - [ ] use an ORM
- [ ] keep the changes such that the datastore can change in the future to other database or store
- [ ] the data is stored as subjects , topics and links.
- [ ] the links will also have an additional set of tags that will further help in categorization if needed.
- [ ] migration script to move the data from markdown files to postgresql.
    - [ ] the filenames becomes the subjects
    - [ ] the subjects currently inside each file becomes the topic
    - [ ] the links will belong to topics and topics belong to some subject.
- [ ] script to initialize the database, tables, users and so on.
- [ ] remove the lines between each row. make it compact
- [ ] put resume on your website url path

### Extras - Not required now
- [ ] New tab which is active tab that contains selected tabs. You can basically keep your entire browser empty and have only one active tab - basically a bookmark manager
