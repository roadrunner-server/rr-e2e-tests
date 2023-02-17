## Test execution

### Run RR

```bash
rr serve
```

### Run client script

```bash
php app.php child
```

## Preparing

### Install latest composer dependencies

```bash
composer update
```

> **Note**
> Use `composer install` to get the locked dependencies.

### Download RoadRunner

```bash
./vendor/bin/rr get -s alpha -f v2023.1.0-alpha.2
```

### Run Temporal Server


