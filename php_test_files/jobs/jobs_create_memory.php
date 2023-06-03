<?php

use Spiral\RoadRunner\Jobs\Jobs;
use Spiral\RoadRunner\Jobs\Queue\MemoryCreateInfo;
use Spiral\RoadRunner\Jobs\Consumer;
use Spiral\RoadRunner\Jobs\Task\ReceivedTaskInterface;

ini_set('display_errors', 'stderr');
require dirname(__DIR__) . "/vendor/autoload.php";

$jobs = new Spiral\RoadRunner\Jobs\Jobs(
    Spiral\Goridge\RPC\RPC::create('tcp://127.0.0.1:6001')
);

$queue = $jobs->create(new MemoryCreateInfo(
    name: 'example',
    priority: 42,
    prefetch: 10,
));
$queue->resume();

$consumer = new Spiral\RoadRunner\Jobs\Consumer();

while ($task = $consumer->waitTask()) {
    $task->complete();
}
