# envgen

_envgen_ is CLI tool that generates .env files for subpackages in your project based on a configuration file.

The purpose of this tool is to:
 - avoid using a bash script: harder to maintain, test, lacks the safety 
 of a statically typed and compiled language like Go;
 - provide flexibility and ease of use/reuse: additional features can be 
 easily added and made available to projects already using the tool.

### Usage

> `$ envgen <vars config file> [env file 1] ... [env file N]`

`$ envgen config.yaml` will read vars already available in the environment

`$ envgen config.yaml .part.env .pck.env` will read vars already available in the environment and will load vars from the passed files

Here's the structure of the configuration file:

```yaml
branchVarName: CIRCLE_BRANCH
branchVarDefault: develop

branches:
  - name: develop
    suffix: _DEV
  - name: staging
    suffix: _STG

packages:
  - package: awesome-crawler
    envFile: .crawler.env
    variables:
      - AC_THRESHOLD
      - AC_TITLE
  - package: web-server/api
    variables:
      - WS_PORT
      - WS_ADDRESS

globals:
  - V_DATABASE
```

Configuration details:

| Key               | Description |
| ----------------- | ------------- |
| branchVarName     | name of the env var that contains the CI branch (CIRCLE_BRANCH), for CircleCI |
| branchVarDefault  | default value of `branchVarName` (optional, defaults to "develop") |
| branches          | list of branches and their suffixes |
| packages          | list of subpackages |
| packages.envFile  | name of the env file to be created (optional) |
| packages.package  | subpackage path to where the .env file will be written |
| packages.variables| subpackage environment variables to generate |
| globals           | list of variables that are environment independent and global to all subpackages |

For each entry in the `packages` array, a `.env` file will be created in the path `package`, 
with the variables defined in `variables` and `globals`.

#### Example:

Project structure:
```
├── bla.go 
├── awesome-crawler/
│   └── index.js
└── web-server/
  └── api/
    └── index.html
```

Loaded env vars:
```
CIRCLE_BRANCH=develop
V_DATABASE=somedburl
WS_ADDRESS_DEV=www.sample.web
WS_PORT_DEV=1234
AC_TITLE_DEV=sometitle 
AC_THRESHOLD_DEV=50
```
 
Resulting structure:
```
├── bla.go 
├── awesome-crawler
│   ├── .crawler.env
│   └── index.js
└── web-server
  └── api/
    ├── .env
    └── index.html
```

Content of awesome-crawler/.crawler.env:
```
V_DATABASE=somedburl
AC_TITLE=sometitle
AC_THRESHOLD=50
```

Content of web-server/.env:
```
V_DATABASE=somedburl
WS_ADDRESS=www.sample.web
WS_PORT=1234
```
