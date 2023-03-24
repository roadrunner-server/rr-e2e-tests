## Test execution

### Run RR

```bash
rr serve
```

### Run client script

```bash
php app.php child
```

> **Note**
> The first two runs will be OK. Worker will be selfkilled each 3rd ChildWorkflow run.

## Preparing

### Install latest composer dependencies

```bash
composer update
```

> **Note**
> Use `composer install` to get the locked dependencies.

### Download RoadRunner

```bash
./vendor/bin/rr get -s alpha -f v2023.1.0-beta.1
```

### Run Temporal Server


