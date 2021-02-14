This repository contains various scripts to collect data about development of open sources projects.

Used mainly to get metrics of Apache projects but can be used for any Github organization.

Most of the subcommands use a simple (file-based) key-value store to get and update the raw data (incrementally) and a
separated command to extract data to csv/parquet files.

Examples:

Downloading Github stats:

```
export GITHUB_TOKEN=...
opm github update repo --org=...

#pr information
opm github update pr
opm github extract pr

#commits
opm github update checkout
opm github extract checkout

```

Downloading mailing list stats:

```
#For private lists use export PONYMAIL_COOKIE=...

#Use filter if you are not interested about all mail lists
opm ponymail getlists --filter=ozone.apache.org

#Download raw data
opm ponymail update

#Extract csv files 
opm ponymail extract 
```

Downloading Jira information

```
opm jira update issue
opm jira extract issue
```