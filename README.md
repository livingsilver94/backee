# Backee
[![Go Report Card](https://goreportcard.com/badge/github.com/livingsilver94/backee)](https://goreportcard.com/report/github.com/livingsilver94/backee)

</br>

Backee is configuration restorer for Unix and Windows computers. It reads a series of `service.yaml` files that contain operating system dependencies, dependencies among other services and POSIX or Powershell scripts (the latter on Windows). Such sections are then used to restore services that a user wanted to backup, right at your fingertip.

It also possible to restore files without scripts. The `links` directory symbolic-links files to their destination path, while the  `data` directory *copies* files, optionally by editing it using the Go template engine, so that a file could be customized for a particular platform on-the-fly. You can think Backee as an advanced dotfiles manager, whilst easy to use with its declarative definition files.

While not strictly necessary, Backee should be run by a privileged user to unleash its potential: restoring system-wide configuration files, webserver resources or installing packages, to name a few.<br/>
On Unix, it links and copies files impersonating the owner of the directory.

If you are using `sudo`, please read the `sudo` caveats below.

## Configuration format

Services to be restored are represented by directories inside a parent directory.

Each service directory contains:

 - A `service.yaml` file, and/or multiple `service_variantName.yaml`. See more documentation below.
 - Optional `data` directory, containing files to process.
 - Optional `links` directory, containing files to be symlinked untouched.

The following is `service.yaml`'s format:

|Key|Type|Meaning|
|---|---|---|
|`depends`|`list(str)`|List of service names as dependencies.</br>The services are then parsed from their directories.|
|`setup`|`str`|Shell or Powershell script executed before packages installation.</br>Doesn't support variables. Data written to the script's stderr is logged on terminal, to show custom messages.|
|`pkgmanager`|`list(str)`|Package manager command with its flags. The package manager must accept a list of package names appended, that will be passed by Backee. Defaults to `["pkcon", "install", "-y"]`.|
|`packages`|`list(str)`|OS packages to install.|
|`links`|`dict(str, str)`|Source-destination pairs for symlinking files, to make Backee act as a dotfile manager. The source path is relative to the `links` directory, while the destination is an absolute, or relative to the current directory, path of the file to link (NOT the directory that will contain it, but the file itself). Non existing parent directories are automatically created and environment variables can be used to define the path: use the `${VAR_NAME}` syntax to match the `VAR_NAME` enviroment variable.|
|`variables`|`dict(str, str)`|Variables passed to the template engine for file copies and the finalization script. Variables can be in clear text or secret.|
|`copies`|`dict(str, str)`|Source-destination pairs for copying files. The source path is relative to the `data` directory, while the destination is an absolute, or relative to the current directory, path of the file to copy (NOT the directory that will contain it, but the file itself). Non existing parent directories are automatically created and environment variables can be used to define the path: use the `${VAR_NAME}` syntax to match the `VAR_NAME` enviroment variable. The file can be edited on-the-fly using the Go template engine, using values in `variables`.|
|`finalize`|`str`|Shell or Powershell script executed as the final stage.</br>It supports variables defined in `variables` using the Go template engine. You may refer to the implicit `datadir` variable to access files inside the `data` directory. Data written to the script's stderr is logged on terminal, to show custom messages.|

Keys are processed in the above order. Each key is optional, to the point it's (pointlessly) possible to write a no-op service.

See `service.example.yaml` for a complete service definition file.

### Secret variables

Backee supports KeepassXC as the secret manager for variables that shouldn't be disclosed. Use the `keepassxc` kind of variable for that. Ensure `keepassxc-cli` is available. See `backee --help` for how to pass the database path, username and password.

## Variants

A service may have different configuration files or scripts depending on the operating system it's being installed on. While the `service.yaml` file contains one-catches-all definitions, a custom `service_customName.yaml` may be written to specialize the definitions for a certain platform. When a custom variant name is passed to the CLI, only `service_customName.yaml` will be parsed.

Example: run `backee --variant homeServer nginx`. This will fetch definitions from `nginx/service_homeServer.yaml`.

## If you are using `sudo`…

Running `sudo` alone doesn't preserve your environment variables. If you rely on them for links and/or copies, or in your setup and finalization scripts, run `sudo -E backee` to keep them.

Some platforms configure `sudo` to not retain your $HOME environment variable: you'll be using the root home instead of yours. Use `sudo -E HOME="$HOME" backee` to ensure your home will be retained.

## License

This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0. If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
