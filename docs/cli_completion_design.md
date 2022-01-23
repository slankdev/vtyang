# CLI Completion Design

## Design Directions

- (1) Full Scratch
	- pros: easy to develop as poc level
	- cons: can't use `grep` and something with cli
- (2) Use Existing Mechanism
	-	pros: can use unix cli tools (`grep`, etc...)
	- cons: I'm not sure how represent it...

## Survey1: current software capability

```bash
$ bind '"?" : "echo slankdev\n"'
$ ?
slankdev
$
```
