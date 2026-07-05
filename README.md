# mfeeder

mfeeder is a small Windows tool that records which desktop windows are open and focused.

It is made of two binaries:

- `mfeederd.exe`: the daemon that watches windows and writes data to SQLite
- `mfeeder.exe`: the CLI used to read the collected data and manage exclusions

The goal is simple: get a local summary of where time was spent during the day.

## Status

This project is currently Windows only.

Data is stored locally in:

```text
%LOCALAPPDATA%\mfeeder\mfeeder.db
```

The config file is:

```text
%LOCALAPPDATA%\mfeeder\mfeeder.conf
```

Logs are stored in:

```text
%LOCALAPPDATA%\mfeeder\logs
```

If you want to use a different data folder, set the `MFEEDER_DATA_DIR` environment variable.

Example:

```text
MFEEDER_DATA_DIR=D:\mfeeder-data
```

In that case the database, config file, and logs are stored there instead of `%LOCALAPPDATA%\mfeeder`.

## Install

Download the latest release from the GitHub Releases page.

The release contains:

```text
mfeeder.exe
mfeederd.exe
```

Create a folder, for example:

```text
C:\mfeeder
```

Put these files inside it:

```text
C:\mfeeder\mfeeder.exe
C:\mfeeder\mfeederd.exe
```

### Configure the daemon in Task Scheduler:

To start it automatically with Windows, use Task Scheduler:

1. Open **Task Scheduler**
2. Click **Create Task**
3. In **General**:
   - name: `mfeederd`
   - enable **Run only when user is logged on**
   - optionally enable **Run with highest privileges** if needed
4. In **Triggers**:
   - add a new trigger
   - begin the task: **At log on**
5. In **Actions**:
   - action: **Start a program**
   - program:
     ```text
     C:\mfeeder\mfeederd.exe
     ```
   - start in:
     ```text
     C:\mfeeder
     ```
6. In **Conditions**:
   - disable conditions you do not want, for example "start only if on AC power"
7. In **Settings**:
   - enable **Allow task to be run on demand**
   - enable **Run task as soon as possible after a scheduled start is missed**
8. Save the task
9. Right click the task and choose **Run** to test it

The `Start in` value is not required for data files anymore, but it is still a good idea to keep it set to the install folder.
The first run creates the default config file if it does not exist.
If you want to run `mfeeder.exe` from any terminal, add this folder to the Windows `Path` environment variable:

```text
C:\mfeeder
```

One way to do it:

1. Open the Windows Start menu
2. Search **Environment Variables**
3. Open **Edit environment variables for your account**
4. Select `Path`
5. Click **Edit**
6. Click **New**
7. Add:
   ```text
   C:\mfeeder
   ```
8. Save everything
9. Open a new terminal

Then check:

```powershell
mfeeder --help
```

## Usage

The examples below assume `C:\mfeeder` is in the Windows `Path`.

Show today's data:

```powershell
mfeeder day
```

Show a specific day, using `MM-DD`:

```powershell
mfeeder day 01-31
```

Group by executable:

```powershell
mfeeder day -e
```

Group by project name, mainly useful for IDE windows:

```powershell
mfeeder day -p
```

Show current exclusions:

```powershell
mfeeder get ex
```

Add an exclusion:

```powershell
mfeeder ex add chrome
```

Remove an exclusion:

```powershell
mfeeder ex rm chrome
```

## Config

The config file contains a comma-separated list of excluded executable names, window titles, or window classes:

```text
EXCLUSIONS=chrome,explorer,WindowsTerminal
```

Excluded windows are not recorded.

## Build From Source

Requirements:

- Go
- Windows

Create the output folder:

```powershell
mkdir dist
```

Build the daemon:

```powershell
go build -ldflags="-H=windowsgui" -o .\dist\mfeederd.exe .\daemon\
```

`-H=windowsgui` prevents Windows from opening a console window when the daemon starts.

Build the CLI:

```powershell
go build -o .\dist\mfeeder.exe .\cli\
```

## Notes

This is a local tool. It does not upload data anywhere.

The binaries can be installed in any stable folder. Runtime data is stored under `%LOCALAPPDATA%\mfeeder`.

## Uninstall

To remove mfeeder:

1. Delete the `mfeederd` task from Task Scheduler
2. Delete the install folder, for example:
   ```text
   C:\mfeeder
   ```
3. Optionally delete local data:
   ```text
   %LOCALAPPDATA%\mfeeder
   ```
