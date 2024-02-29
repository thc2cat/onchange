# Onchange

On change, do something.

Monitor files and when modifications occurs launch a command.

Commands are executed with "bash -c" to simplify code.

## Exemple

```shell
(master) $ onchange "make update"
14:27:36 - make update
```

For more complex behavior like stop/rebuild/live reload check <https://github.com/cosmtrek/air>
