<!---
SPDX-FileCopyrightText: Fabio Forni <development@redaril.me>
SPDX-License-Identifier: CC0-1.0
-->

# Backee
[![Go Report Card](https://goreportcard.com/badge/github.com/livingsilver94/backee)](https://goreportcard.com/report/github.com/livingsilver94/backee)

</br>

> ⚠️ Windows support is untested at the moment, and thus considered experimental.

Backee is configuration restorer for Unix and Windows computers. It reads a series of `service.yaml` files that contain operating system dependencies, dependencies among other services and POSIX or Powershell scripts (the latter on Windows). Such sections are then used to restore services that a user wanted to backup, right at your fingertip.

It also possible to restore files without scripts. The `links` step symbolic-links files to their destination path, while the  `copies` step *copies* files, optionally by editing them using a template engine, so that a file could be customized for a particular user or platform on-the-fly. You can think of Backee as an advanced dotfiles manager, whilst easy to use with its declarative definition files.

Backee performs operations as the user that run it by default. On UNIX, If a permission is denied while copying or linking files, it retries by calling a privilege elevation utility. `sudo` and `doas` are supported at the moment.

See it in action!

[![asciicast](https://asciinema.org/a/hMEVqtjppGfROguT00eZFdyTQ.svg)](https://asciinema.org/a/hMEVqtjppGfROguT00eZFdyTQ)

## Configuration format

Services to be restored are represented by directories inside a parent directory.

Each service directory contains:

 - A `service.yaml` file, and/or multiple `service_variantName.yaml`. See further documentation on variants below.
 - Optional `data` directory, containing files to process.
 - Optional `links` directory, containing files to be symlinked untouched.

As an example, this directory tree contains two services named mysql and nginx. mysql contains the optional data directory to eventually restore *my.conf*.

<pre>
.
├── <b>mysql</b>
│   ├── data
│   │   └── my.conf
│   └── service.yaml
└── <b>nginx</b>
    └── service.yaml
</pre>

The following is `service.yaml`'s format in the basic form. Some keys support an extended format for more flexibility: see [`service.example.yaml`](service.example.yaml) to learn more.

|Key|Type|Meaning|
|---|---|---|
|`depends`|`list(str)`|List of service names as dependencies.</br>These services will be installed first.|
|`setup`|`str`|Shell or Powershell script executed before packages installation.</br>Doesn't support variables. Data written to the script's stderr is logged on terminal, to show custom messages.|
|`pkgmanager`|`list(str)`|Package manager command with its flags. The package manager must accept a list of package names appended, that will be passed by Backee. Defaults to `["pkcon", "install", "-y"]`.|
|`packages`|`list(str)`|OS packages to install.|
|`links`|`dict(str, str)`|Source-destination pairs for symlinking files/directories. The source path is relative to the service's `links` directory, while the destination is the symlink path. Non existing parent directories are automatically created. Variables can be used to compose the destination path.|
|`variables`|`dict(str, str)`|Extra variables on top of environment variables.|
|`copies`|`dict(str, str)`|Source-destination pairs for copying files. The source path is relative to the service's `data` directory, while the destination is the path of the file copied. Non existing parent directories are automatically created. Variables can be used to compose the destination path and to customize the content of each file.|
|`finalize`|`str`|Shell or Powershell script executed as the final stage.</br>It supports variables to customize the script. You may also refer to the implicit `datadir` variable to access files inside the `data` directory. Data written to the script's stderr is logged on terminal, to show custom messages.|

Keys are processed in the above order. Each key is optional, to the point it's (pointlessly) possible to write a no-op service.

### Secret variables

Backee supports KeepassXC as the secret manager for variables that shouldn't be disclosed. Use the `keepassxc` kind of variable for that. Ensure `keepassxc-cli` is available. Run `backee install --help` to learn how to pass the database path, username and password.

## Variants

A service may have different configuration files or scripts depending on the operating system it's being installed on. While the `service.yaml` file contains one-catches-all definitions, a custom `service_customName.yaml` may be written to specialize the definitions for a certain platform. When a custom variant name is passed to the CLI, only `service_customName.yaml` will be parsed.

Example: run `backee --variant homeServer nginx`. This will fetch definitions from `nginx/service_homeServer.yaml`.

## License

This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0. If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
