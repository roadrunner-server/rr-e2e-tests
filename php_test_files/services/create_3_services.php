<?php
use Spiral\RoadRunner\Services\Manager;
use Spiral\RoadRunner\Services\Exception\ServiceException;
use Spiral\Goridge\RPC\RPC;

ini_set('display_errors', 'stderr');
require dirname(__DIR__) . "/vendor/autoload.php";

$rpc = RPC::create('tcp://127.0.0.1:6001');
$manager = new Manager($rpc);


try {
    $result = $manager->create(
        name: 'listen-jobs', 
        command: 'sleep 5',
        processNum: 3,
        execTimeout: 15,
        remainAfterExit: false,
        env: ['APP_ENV' => 'production'],
        restartSec: 30
    );
    
    if (!$result) {
        throw new ServiceException('Service creation failed.');
    }
} catch (ServiceException $e) {
    // handle exception
}
