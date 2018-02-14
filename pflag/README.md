# Compiling dbdeployer
If you want to compile dbdeployer, you must also modify this file, which is a dependency of the cobra package.

    github.com/spf13/pflag/string_slice.go

The reason is that encoder/csv uses a literal value to initialize the field delimiter in readers and writers, and the depending package does not offer an interface to modify such value (comma ',').

The problem that this creates is related to multi-string flags in dbdeployer. By default, due to cobra using the encoding/csv package, if the value contains a comma, it is assumed to be a separator between values. This may not be the case with dbdeployer, where we can pass option-file directives (which sometimes contain commas) and SQL commands (which **often** contain commas.)

Whith this change, the default field separator is a semicolon.
