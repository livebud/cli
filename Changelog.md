# 0.0.13 / 2025-01-20

- add support for using environment variables
- usage now shows defaults and optionals
- better error messages for args.

# 0.0.12 / 2025-01-12

- fix bug where flags were sometimes getting shared on subcommands

# 0.0.11 / 2024-12-07

- support passing flags out of order

# 0.0.10 / 2024-10-26

- fix help message coloring

# 0.0.9 / 2024-10-26

- support basic tab completion using `COMP_LINE` and `complete -o nospace -C <cmd> <cmd>`

# 0.0.8 / 2024-10-26

- **BREAKING:** switch away from colon-based subcommands.
  You can still use colon-based commands, you just use them as commands like `cli.Command("fs:cat", ...)` and don't nest.
- add ability to find and modify commands

# 0.0.7 / 2024-05-04

- better error message for enums

# 0.0.6 / 2024-05-04

- add help message back in for args (not used yet)

# 0.0.5 / 2024-05-04

- add enum support
- Remove unused add ability to infer flags and args from struct tags
- Remove unused support calling multiple runners at once

# 0.0.4 / 2023-10-22

- update release script

# 0.0.3 / 2023-10-22

- Add ability to infer flags and args from struct tags.
- Support calling multiple runners at once

# 0.0.2 / 2023-10-07

- switch to using signal.NotifyContext

# 0.0.1 / 2023-06-21

- Initial release
