# Changelist of gmig releases

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