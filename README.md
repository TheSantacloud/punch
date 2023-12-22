# Punch

A simple CLI punchcard application

## TODO

- [ ] add sync support - pull
- [ ] add sync support - push
- [ ] add autosync on dedicated actions (start, finish, edit, remove)
- [ ] make cli get only for sessions, and separate out companies
- [ ] add report --all-companies
- [ ] delete day (via delete and via edit)
- [ ] add `mockgen -source=pkg/repositories/interfaces.go -destination=tests/mock.go -package=tests` to the pre-commit hook
- [ ] add tests
- [ ] spreadsheet init (sync setup)
- [ ] add github actions
- [ ] write a complete readme with guides
- [X] support multi work periods within one day (and rename day session)
- [X] refactor `Database` completely. it's horrible. make it a repository
- [X] add named remotes with types (only with sheets available)
- [X] change sync to push/pull remotes
- [X] add comment for a workday (-m)
- [X] edit company
- [X] currency per company rather than general
- [X] make week month and year string flags with default values of current
- [X] add monthly/weekly report by company/all
- [X] don't do anything (sync) when edit buffers weren't changed
- [X] sync on `edit` command
- [X] add retroactive punchcard reporting 
- [X] derive PPH from the spreadsheet
- [X] ability to add/edit companies and their PPH
- [X] add more than one company support

