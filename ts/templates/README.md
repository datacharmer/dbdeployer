# Templates for testdata files

Every `.gotxt` file is a template that will eventually become a test file in `./testdata`.

Each file will be treated as a `text/template` and converted into a `.txt` after each `{{.Label}}` has been converted into the proper value.

