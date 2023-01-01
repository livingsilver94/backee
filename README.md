# Backee

Backee is configuration restorer for Unix computers. It reads a series of `service.yaml` files that contain operating system dependencies, dependencies among other services and POSIX shell scripts. Such sections are then used to restore services that a user wanted to backup. It also possible to restore files provided in the `data` directory. You can think it as a dotfiles manager, but more powerful and with the ability to restore system-wide configuration.

Backee executes all child processes as the user executing Backee, so you may need to add privilege elevation commands where necessary such as `sudo`.

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
|`setup`|`str`|Shell script executed before packages installation.</br>Doesn't support variables. Data written to stderr is logged by confrestorer, to show custom messages.|
|`pkgmanager`|`list(str)`|Package manager command an its flags. The package manager must accept a list of package names appended, that will be passed by `confrestorer`. Defaults to `["pkcon", "install", "-y"]`.|
|`packages`|`list(str)`|OS packages to install.|
|`links`|`dict(str, str)`|Source-destination pair for symlinking files, to make confrestorer act as a dotfile manager. The source path is relative to the `links` directory, while the destination path is an absolute or relative path of the file to link (NOT the directory that will contain it, but the file itself). Non existing parent directories are automatically created and environment variables can be used to define the path: use the `${VAR_NAME}` syntax to match the `VAR_NAME` enviroment variable.|
|`vars`|`dict(str, str)`|Variables available in the finalization script.|
|`finalize`|`str`|Shell script executed after links installation.</br>Supports variables defined in `vars`, enclosed between `%` (e.g. `%myvar%`). Data written to stderr is logged by confrestorer, to show custom messages.|

Keys are processed in the above order.

Each key is optional, to the point it's (pointlessly) possible to write a no-op service.

## Variants

A service may have different configuration files or scripts depending on the operating system it's being installed on. While the `service.yaml` file contains one-catches-all definitions, a custom `service_customName.yaml` may be written to specialize the definitions for a certain platform. When a custom variant name is passed to the CLI, only `service_customName.yaml` will be parsed.

Example: run `backee -variant homeServer nginx`. This will fetch definitions from `nginx/service_homeServer.yaml`.
