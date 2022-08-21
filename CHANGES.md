# Changelist of gmig releases

## v1.17.0 [2022-08-21]

- fix crash when JSON config is invalid
- update dependencies (also for vulnerability)

## v1.16.0 [2021-08-18]

- fix handling `migrations` flag in force commands

## v1.15.0 [2020-05-13]

- add util for adding patch rules to a loadbalancer path matcher
- improved error reporting

## v1.14.0 [2020-03-13]

- add support for conditional migrations using boolean expression

## v1.13.0 [2020-03-11]

- show status after up,down and force state

## v1.12.2

- fix bug introduced with 1.12.0/1 when saving state in GCS

## v1.12.0

- fix handling failure when force do has invalid config
- improve printing error message
- last migration state is now temporary stored in OS temp directory

## v1.11.0

- added functionality to quickly get all environment variable from a gmig config and also $PROJECT, $REGION and $ZONE

## v1.10.6

- better error message when migration is not found
- fix panic when config folder is misspelled or missing
- fix panic when force command is wrongly used

## v1.10.4

- simplify list, fix export iam policy
- use yaml v2
- fix handling invalid config folder

## v1.10.1

- fix warning message, remove vendor folder in favor of go modules.

## v1.10.0

- switch to YAML for configuration (JSON is fallback)

## v1.9.0

- add the "template" command for simple configuration transformation.

## v1.8.2

- fixes bug in collecting migration files: it should not recurse into subdirectories

## v1.8.0

- replace timestamps in migration files by indices. (010,015,...)


## v1.7.0, 2018-09-28

- add "plan" command that logs all commands that would be executed on "up".

## v1.6.0

- add options to new to set the commands for the do,undo,view section directly
- use verbose flag to echo expanded environment variables in commands

## v1.5.0

- add "view" command to show the status of infrastructure, per migration.

## v1.4.0

- add util functions to update named-ports for an instance group

## v1.3.0

- added --migrations option for up,down,status,force

## v1.2.2, 2018-08-24

- all commands in a section (do or undo) from a migration is now executed using a temporary shell script.
  this allows for using shell variables that can be used by other commands within the section.

## v1.2.0, 2018-08-22

- the up command can now optionally stop after a specified migration filename.
- initial configuration has sample environment variable.
- better argument usage documentation.

## v1.1.1, 2018-08-20

- improve error report when YAML file is not valid.

## v1.1.0, 2018-08-12

- support `.yml` extension for migration files.

## v1.0.0

- initial release