use Spiral\Goridge\RPC\RPC;
use RoadRunner\Logger\Logger;

$rpc = RPC::fromEnvironment(new \Spiral\RoadRunner\Environment([
    'RR_RPC' => 'tcp://127.0.0.1:6001'
]));

$logger = new Logger($rpc);

/**
 * debug mapped to RR's debug logger
 */
$logger->debug('Debug message');

/**
 * error mapped to RR's error logger
 */
$logger->error('Error message');

/**
 * log mapped to RR's stderr
 */
$logger->log("Log message \n");

/**
 * info mapped to RR's info logger
 */
$logger->info('Info message');

/**
 * warning mapped to RR's warning logger
 */
$logger->warning('Warning message');