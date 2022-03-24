<?php

use Nyholm\Psr7\Factory\Psr17Factory;
use Nyholm\Psr7\Response;
use Nyholm\Psr7\Stream;
use Spiral\RoadRunner;

ini_set('display_errors', 'stderr');
require __DIR__ . "/vendor/autoload.php";

$worker = RoadRunner\Worker::create();
$psr7 = new RoadRunner\Http\PSR7Worker(
    $worker,
    new Psr17Factory(),
    new Psr17Factory(),
    new Psr17Factory()
);

$psr7->chunk_size = 10 * 10 * 1024;
$filename = './file.tmp';

while ($req = $psr7->waitRequest()) {
    try {
        $fp = \fopen($filename, 'rb');
        \flock($fp, LOCK_SH);
        $resp = (new Response())->withBody(Stream::create($fp));
        $psr7->respond($resp);
    } catch (\Throwable $e) {
        $psr7->getWorker()->error((string)$e);
    }
}