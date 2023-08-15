<?php

declare(strict_types=1);

namespace Temporal\Tests\Workflow;

use Carbon\CarbonInterval;
use Temporal\Client\GRPC\ServiceClient;
use Temporal\Client\WorkflowClient;
use Temporal\Client\WorkflowOptions;
use Temporal\Workflow\WorkflowRunInterface;

require_once __DIR__ . '/vendor/autoload.php';

$host = 'localhost:7233';

$workflowClient = WorkflowClient::create(ServiceClient::create($host));

$_SERVER['VAR_DUMPER_FORMAT'] = 'server';
$_SERVER['VAR_DUMPER_SERVER'] = '127.0.0.1:9912';

function findSkipped(array $result, int $waves): array
{
    // Get numbers from result
    $numbers = [];
    foreach ($result as $item) {
        // Regexp version
        \preg_match('/Name-\d+-(\d++)/', $item, $matches);
        $numbers[] = (int)$matches[1];
    }
    // Find skipped numbers
    $skipped = [];
    for ($i = 0; $i < $waves; $i++) {
        if (!\in_array($i, $numbers, true)) {
            $skipped[] = $i;
        }
    }

    return $skipped;
}
/*
 * @return array{WorkflowRunInterface, SimpleWorkflow}
 */
function runWorkflow($workflowClient): array
{
    $workflow = $workflowClient->newWorkflowStub(
        SimpleWorkflow::class,
        WorkflowOptions::new()->withWorkflowExecutionTimeout(CarbonInterval::minutes(25)),
    );

    echo "Starting SignalWorkflow...\n";
    $run = $workflowClient->start($workflow);

    echo \sprintf(
        "Started: WorkflowID=\e[35m%s\e[0m, RunID=\e[35m%s\e[0m\n",
        $run->getExecution()->getID(),
        $run->getExecution()->getRunID(),
    );

    return [$run, $workflow];
}


/// CONFIG ///
$workflowsCount = 20;
$waves = 120;
//////////////

$time['begin'] = \microtime(true);

/** @var array<array{WorkflowRunInterface, SimpleWorkflow}> $runs */
$runs = [];

for ($i = 0; $i < $workflowsCount; $i++) {
    $runs[] = runWorkflow($workflowClient);
}
$time['workflows were run'] = \microtime(true);

$iterator = 0;
for ($i = 0; $i < $waves; $i++) {
    // Print wave number
    echo \sprintf(" \e[32m%d\e[0m", $i + 1);

    foreach ($runs as $key => $run) {
        $name = "Name-$key-$i--" . $iterator++;
        $run[1]->addName($name);
    }
}
$time['signals were sent'] = \microtime(true);

echo "\n";

foreach ($runs as $key => $run) {
    echo \sprintf("Send exit [wf:%d]... ", $key);
    try {
        $run[1]->exit();
    } catch (\Throwable $e) {
        echo \sprintf("\e[31mSIGNAL ERROR\e[0m\n");
        echo \sprintf("Error: \e[31m%s\e[0m\n", $e->getMessage());
        // more details
        echo \sprintf("previous: \e[31m%s\e[0m\n", $e->getPrevious()?->getMessage());
        echo \sprintf("previous previous: \e[31m%s\e[0m\n", $e->getPrevious()?->getPrevious()?->getMessage());

        continue;
    }
    // Out result
    try {
        $result = $run[0]->getResult();
    } catch (\Throwable $e) {
        echo \sprintf("\e[31mGET RESULT ERROR\e[0m\n");
        echo \sprintf("Error: \e[31m%s\e[0m\n", $e->getMessage());
        continue;
    }

    $time['first result'] = \microtime(true);

    if (\count($result) === $waves) {
        // Print Ok
        echo \sprintf("Got count: \e[32m%d\e[0m\n", \count($result));
    } else {
        // Print Error
        echo \sprintf("Result [wf:%d]: \e[31mERROR\e[0m\n", $key);
        echo \sprintf("Got count: \e[31m%d\e[0m instead of \e[32m%d\e[0m\n", \count($result), $waves);
        echo \sprintf("Skipped names: \e[31m%s\e[0m\n", \implode(', ', findSkipped($result, $waves)));
    }
}
$time['end'] = \microtime(true);

$prev = \reset($time);
foreach ($time as $label => $s) {
    echo \sprintf("%s: \e[32m%.4f\e[0m\n", $label,  $s - $prev);
    $prev = $s;
}

