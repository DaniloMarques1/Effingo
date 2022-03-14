## Effingo

Traversy a given file system and finds duplicate files.

### Usage

To find and print the duplicate files names. It will print the full path to the
file (including its name).

```console
effingo --help
effingo -d /path
```

If you would like to remove the duplicate files, add a `-r` flag

```console
effingo -d /path -r
```

By default the search does not include hidden files (dotfiles), if you would
like to include them, you need to provide the `-a` flag.

```console
effingo -d /path -a -r
```
