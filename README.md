# Comic Book Zip Archive Tool

`cbz` is a command-line tool for viewing and manipulating `ComicInfo.xml` files in comic book zip archive files. The schema for `ComicInfo.xml` files is defined by the [Anansi Project](https://github.com/anansi-project/comicinfo).

## Why?

This tool exists solely so I can add `AgeRating` metadata to comic CBZ files so that my family's private  [Kavita Reader](https://www.kavitareader.com/) instance doesn't present unsuitable content to young individuals. However, for fun I've added support for updating other metadata fields.

## Limitations

Does not currently support setting fields deeper than the first level, such as elements under the `Pages` array.

## Running

### Viewing Metadata

View a comic archive's `ComicInfo.xml` file:

```shell
cbz show comic.cbz
```

Note that the `show` command can only operate on a single comic archive.

### Set Metadata

Set an `AgeRating` on a comic files:

```shell
cbz set AgeRating=Teen comic.cbz
```

Set multiple metadata fields:

```shell
cbz set AgeRating=Teen "Writer=Terry Moore" comic.cbz
```

The `set` command can operate on multiple comic archives:

```shell
cbz set AgeRating=Teen "Writer=Terry Moore" series1/comic1.cbz series2/*.cbz
```

The `import` command "imports" a CBR file by converting it into a CBZ file.

```shell
cbz import comic.cbr
ls
comic.cbr comic.cbz
```