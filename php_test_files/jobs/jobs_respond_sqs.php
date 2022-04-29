<?php

/**
 * @var Goridge\RelayInterface $relay
 */

use Spiral\Goridge;
use Spiral\RoadRunner;
use Spiral\Goridge\StreamRelay;

ini_set('display_errors', 'stderr');
require dirname(__DIR__) . "/vendor/autoload.php";

$rr = new RoadRunner\Worker(new StreamRelay(\STDIN, \STDOUT));

while ($in = $rr->waitPayload()) {
    try {
        $ctx = json_decode($in->header, true);
        $headers = $ctx['headers'];

        $rr->respond(new RoadRunner\Payload(json_encode([
            'type' => 2,
            'data' => [
                'queue' => 'https://sqs.us-east-1.amazonaws.com/588160034479/resp-queue',
                'payload' => 'foo'
            ]
        ])));
    } catch (\Throwable $e) {
        $rr->error((string)$e);
    }
}