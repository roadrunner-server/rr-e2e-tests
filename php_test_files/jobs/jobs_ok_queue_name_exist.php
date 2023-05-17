<?php

/**
 * @var Goridge\RelayInterface $relay
 */

use Spiral\Goridge;
use Spiral\RoadRunner;
use Spiral\Goridge\StreamRelay;
use Spiral\RoadRunner\Jobs\Consumer;
use Spiral\RoadRunner\Jobs\Serializer\JsonSerializer;
use RuntimeException;

ini_set('display_errors', 'stderr');
require dirname(__DIR__) . "/vendor/autoload.php";

$rr = new RoadRunner\Worker(new StreamRelay(\STDIN, \STDOUT));
$consumer = new Consumer($rr, new JsonSerializer);

while ($task = $consumer->waitTask()) {
    try {
        if ('unknown' === $task->getQueue()) {
            throw new RuntimeException('Queue name was not found');
        }

        $task->complete();
    } catch (\Throwable $e) {
        $rr->error((string)$e);
    }
}