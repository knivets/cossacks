# Cossacks test assignment

Generates and logs fibonacci numbers, includes optional encryption.

Requires Go 1.15

## Usage

`generator/generator.go`

```
Usage of generator/generator.go:
  -debug
    	enable debug mode
  -generation_speed int
    	throttle output in numbers per second (default 100)
```

`logger/logger.go`

```
Usage of logger/logger.go:
  -buffer_size int
    	path to file where to store the logs (default 131072)
  -debug
    	enable debug mode which expects --file_path and --log_key to decrypt the contents of the file
  -file_path string
    	path to file where to store the logs
  -flow_speed int
    	control the speed of reading from input (default 100)
  -log_key string
    	encryption key
```

## Examples

Generate fibonacci numbers at a 100/s rate, pipe them to logger which reads at 1000/s speed, encrypts and logs the data into a file:

```
go run generator/generator.go --generation_speed=100 | go run logger/logger.go --log_key=aaaa --file_path=out.log --flow_speed=1000
```
