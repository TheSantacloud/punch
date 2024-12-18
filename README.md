# Punch

A simple CLI punchcard program.

It's a highly configurable, local-first, with remote syncing support. Designed with a UX similar to `kubectl` or `git`.

I developed it for myself, and thought that this might benefit other people so I made it a bit more ready.

## Installation

To install Punch, clone the repository and run `make` in the project directory:

```
git clone git@github.com:dormunis/punch.git
cd punch
make
```

## Usage

### Starting and Ending Work Sessions
- **Punch Toggle**: Simply entering `punch` toggles the work session. If a session is not started, it starts one. If a session is ongoing, it ends it. This functionality assumes you have a `default_client` set in your configuration.

  Toggle start/end session:
  ```bash
  punch
  punch -c Acme     # if there's no default_client set
  ```

- **Explicit Start and End**: You can also start and end sessions explicitly. You can also add a date/time explicitly to it.

  ```bash
  punch start        # start session
  punch start [date] # start session at [date]
  punch start -- -2m    # start session 2 minutes ago
  punch end          # end session
  punch end [date]   # end session at [date]
  punch end -- -10m     # end session 10 minutes ago
  ```

  Supported relative times are:
  - s - second
  - m - minute
  - h - hour
  - d - day
  - M - month
  - y - year

### Get Command
- **Retrieve Client or Session Details**: Use the `get` command to fetch details about clients or work sessions.

  Get details of a specific client:
  ```bash
  punch get client [client_name]
  ```

  Get details of a work session:
  ```bash
  punch get session -c Acme         # get latest session from Acme
  punch get session [date] -v       # get sessions from [date] with verbose information
  punch get session --week -v       # get verbose sessions from last day/week/month/year
  punch get session -- -1.5w        # get all sessions from the past 1.5 weeks
  punch get session -3 -c Acme      # get last 3 sessions from Acme
  punch get session --all -v -o csv # get verbose information in CSV format
  ```

### Add Command
- **Add New Clients or Sessions**: Use the `add` command to add new clients.

  Add a new client:
  ```bash
  punch add client [client_name] [hourly_rate] # currency default to the configured custom_currency
  punch add client [client_name] [hourly_rate] --currency EUR
  ```

### Delete Command
- **Delete Clients or Sessions**: Use the `delete` command to remove clients or sessions.

  ```bash
  punch delete client [client_name]
  punch delete session [session_id]
  ```

### Edit Command
- **Edit Client or Session Information**: Use the `edit` command to modify details of clients or sessions.

  ```bash
  punch edit client [client_name]
  punch edit session --all
  punch edit session [session_id]
  punch edit session --all
  ```

### Additional Tips
- **Setting a Default Client**: For the `punch` toggle feature to work seamlessly, set a default client in your `config.toml`. This eliminates the need to specify a client each time you start a session.
- **Setting a Default Currency**: Currency is set whenever you add a new client, you can bypass it by setting a `default_currency` in the `config.toml`
- **Config easy access**: use `punch config` to access the config file.
- **Help and Command Options**: For more detailed usage of each command, you can use the `--help` option with any command to get additional information and options available.

## Configuration

Punch uses a TOML based configuration

Punch configuration can be accessed by using `punch config` or in `~/.punch/config.toml`.

A default configuration will be generated for you in the aforementioned directory, but
`punch` can use a different config file given the `PUNCH_CONFIG_FILE` environment variable.

### General Structure

The configuration has 3 primary sections:

1. **Settings**: General settings for the application.
2. **Database**: Configuration for the database connection.
3. **Remotes**: Settings for remote synchronization.

### Settings

| Field            | Description                                             | Example             |
|------------------|---------------------------------------------------------|---------------------|
| `editor`         | Text editor for editing purposes (defaults to `vi`)     | `vi`                |
| `currency`       | Default currency for billing. (defaults to USD)         | `USD`               |
| `default_remote` | Default remote for synchronization.                     | `myRemote`          |
| `default_client` | Default client for sessions.                            | `Acme Corp`         |
| `autosync`       | Events triggering auto-sync (start, end, edit, delete). | `["end", "edit"]`   |

Example:
```toml
[settings]
editor = "vi"
default_currency = "USD"
default_remote = "myRemote"
default_client = "Acme Corp"
autosync = ["end", "edit"]
```

### Database

There are only 1 availble database support (for internal use) right now, which is `sqlite3`.

Unless otherwise mentioned, the path of the database is within the configuration directory.

| Field   | Description                                   | Example            |
|---------|-----------------------------------------------|--------------------|
| `Engine`| Database engine (currently supports sqlite3). | `sqlite3`          |
| `Path`  | Path to the database file.                    | `/path/to/punch.db`|

Example:
```toml
[database]
engine = "sqlite3"
path = "/path/to/punch.db"
```

### Remotes

For each remote, you'll define a key and specify its details. Currently, Punch only supports `spreadsheets` (Google Spreadsheet).

#### SpreadsheetRemote

| Field                      | Description                                     | Example                         |
|----------------------------|-------------------------------------------------|---------------------------------|
| `spreadsheet_id`           | ID of the spreadsheet for synchronization.      | `1A2b3C4d5E6f`                  |
| `sheet_name`               | Name of the sheet within the spreadsheet.       | `Sheet1`                        |
| `service_account_json_path`| Path to the service account JSON for access.    | `/path/to/service-account.json` |
| `columns`                  | Define column names for ID, Client, Date, etc.  | See below                       |

Example:
```toml
[remotes.myRemote]
type = "spreadsheet"
spreadsheet_id = "1A2b3C4d5E6f"
sheet_name = "Sheet1"
service_account_json_path = "/path/to/service-account.json"

[remotes.myRemote.columns]
id = "ID"
client = "Client"
date = "Date"
start_time = "Start Time"
end_time = "End Time"
total_time = "Total Time"
note = "Note"
```

## Remotes

Remotes are dedicated for syncing purposes and backups. They are completely optional.

### Google Spreadsheets

1. Using [Google Developer Console](https://console.cloud.google.com/) create a new project and name it whatever you like.
2. Add [Sheets API](https://console.cloud.google.com/apis/library/sheets.googleapis.com) support
3. Create a [Service Account](https://console.cloud.google.com/apis/credentials) and download the JSON
4. Place the Service Account JSON wherever you like (I recommend putting it in `~/.punch/`)
5. Set `service_account_json_path` in the configuration to the path of your service account json.
   (It automatically set to `~/.punch/service-account.json` by default)
6. Create a Spreadsheet in your Google Account, and create a header with the following columns: (you can name them however you like)
    - ID
    - Client
    - Date
    - Start Time
    - End Time
    - Total Time
    - Note
7. (Optional) format the columns with the relevant formats (date, duration, currency, etc)
8. Configure your remote spreadsheet and map the column names to the relevant IDs ([See example](#spreadsheetremote) )
9. Share the google sheet you've created with the service account email within JSON generated

