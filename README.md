# gitall

**gitall** is a CLI tool written in Go that tracks multiple Git repositories in a plain `.db` file and lets you run `git` commands on all of them at once.

It’s not magic. It just works the way I want it to. I don't really expect anyone else to use this honsetly

---

## Features

* Tracks repositories in `~/.gitall.db` (or custom file via `-db`)
* Uses `.gitall.db.sha256` to verify DB integrity
* Runs `git` commands across all repos in the DB
* Logs last 50 operations to `.gitall.db.log`
* Excludes selected paths via `-exclude`
* Custom command execution with streaming stdout/stderr
* Verifies `.git` folder before running any operation

---

## Building gitall

This is a portable Go program. Just clone and build it. Works on Linux and Windows (but really, use Linux).

```
git clone https://github.com/VPeti11/gitall.git
cd gitall
go build -o gitall mian.go
```

You can rename it to whatever you want. I use `gitall` and put it somewhere in my `$PATH`.

---

## Usage

Run `gitall` with one or more flags. If you run it without flags, you’ll get the help menu.

### Commands

| Flag                 | Description                                   |
| -------------------- | --------------------------------------------- |
| `-addrepo <path>`    | Add a Git repo to the DB                      |
| `-delrepo <path>`    | Remove a repo from the DB                     |
| `-listrepos`         | List all repos in the DB                      |
| `-listlastoperation` | Show the last 50 logged commands              |
| `-reinit`            | Clear the DB, log, and checksum (start fresh) |
| `-command <args>`    | Run `git <args>` in every repo                |
| `-exclude <paths>`   | Comma-separated list of repo paths to skip    |
| `-db <file>`         | Use a custom database file instead of default |

All flags before `-command` are parsed. Everything after `-command` is passed to Git directly.

---

### Examples

```
# Add a repo to tracking
./gitall -addrepo ~/projects/website

# Run git status across all tracked repos
./gitall -command status

# Pull all, but skip one directory
./gitall -exclude ~/projects/private -command pull

# Reset everything
./gitall -reinit

# Use custom DB and log files
./gitall -db ~/.gitwork.db -command fetch --all
```

---

## Notes

* This tool will check that Git is installed at launch.
* It reads the SHA256 hash of the DB before any command and refuses to proceed if tampered.
* It checks for a `.git` directory before running Git commands.
* Git output is printed live to your terminal, so no Go nonsense.

---

## License

This software is licensed under the GNU General Public License (V3). This means:

GPLv3 is a free software license that ensures users' freedom to run, study, share, and modify the software.
Any redistributed versions must also be licensed under GPLv3, preserving these freedoms.
You must make the source code available when distributing the program or any derivatives.
You may not impose further restrictions that conflict with the license.
The license also includes protections against tivoization, software patents, and anti-circumvention laws.

| Feature                    | Description                                                                |
| -------------------------- | -------------------------------------------------------------------------- |
| Freedom to Use             | You can run the program for any purpose                                    |
| Freedom to Modify          | You can study the source code and make changes                             |
| Freedom to Share           | You can distribute original or modified versions under the same license    |
| Source Code Disclosure     | When distributing, you must provide or offer access to the source code     |
| No Tivoization / DRM Locks | You cannot use hardware restrictions to block modified software            |
| Patent Protection          | Users are protected from patent claims that would restrict software rights |



You can read the license [here](LICENSE.md)


---

