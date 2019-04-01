Miscellaneous command-line tools
================================

*	[`walk` walks a filesystem hierarchy][walk].

*	[`stest` is a streaming (or filter) version of `test`(1)][stest].

*	[`acme/acmeeval` evaluates Acme commands from inside Acme][acmeeval].

*	[`acme/acmepipe` replaces the body of the current Acme window][acmepipe].

*	[`acme/autoacme` executes a command each time Acme logs an event][autoacme].

I use `autoacme` with scripts that
[apply editorconfig settings when I load][acme-editorconfig] and
[format my source files when I save][acme-autoformat].
These scripts in turn use `acmeeval` and `acmepipe`.

[walk]: https://godoc.org/github.com/mkhl/cmd/walk
[stest]: https://godoc.org/github.com/mkhl/cmd/stest
[acmepipe]: https://godoc.org/github.com/mkhl/cmd/acme/acmeeval
[acmepipe]: https://godoc.org/github.com/mkhl/cmd/acme/acmepipe
[autoacme]: https://godoc.org/github.com/mkhl/cmd/acme/autoacme
[acme-autoformat]: https://gist.github.com/mkhl/69e2be41bfeccb368b52818ebd7f535b#file-acme-autoformat
[acme-editorconfig]: https://gist.github.com/mkhl/5e4cda4f9a262f432eacd592aba5fd54#file-acme-editorconfig
