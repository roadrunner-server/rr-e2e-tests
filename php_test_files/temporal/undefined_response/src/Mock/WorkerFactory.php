<?php

declare(strict_types=1);

namespace Temporal\Samples\Mock;

use Temporal\Worker\Transport\HostConnectionInterface;
use Temporal\Worker\Transport\RoadRunner;

/**
 * It just dies during 3nd ExecuteChildWorkflow
 */
class WorkerFactory extends \Temporal\WorkerFactory
{
    public function run(HostConnectionInterface $host = null): int
    {
        // hack to use private methods
        $fn = function (HostConnectionInterface $host = null) {
            $host ??= RoadRunner::create();
            $this->codec = $this->createCodec();

            while ($msg = $host->waitBatch()) {
                try {
                    $send = $this->dispatch($msg->messages, $msg->context);

                    $host->send($send);

                    static $counter = 0;
                    if (\str_contains($send, 'ExecuteChildWorkflow')) {
                        if (++$counter > 2) {
                            dump('!! SELFKILL !!');
                            die;
                        }
                    }
                } catch (\Throwable $e) {
                    $host->error($e);
                }
            }

            return 0;
        };
        return \Closure::bind($fn, $this, parent::class)($host);
    }
}
