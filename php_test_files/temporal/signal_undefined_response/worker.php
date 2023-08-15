<?php

declare(strict_types=1);

namespace Temporal\Tests\Workflow;

use Temporal\Worker\WorkerOptions;

require_once __DIR__ . '/vendor/autoload.php';
require_once __DIR__ . '/SimpleActivity.php';
require_once __DIR__ . '/SimpleWorkflow.php';

$factory = \Temporal\WorkerFactory::create();
$worker = $factory->newWorker('default', WorkerOptions::new()->withMaxConcurrentActivityExecutionSize(8));

$worker->registerWorkflowTypes(SimpleWorkflow::class);
$worker->registerActivityImplementations(new SimpleActivity());

$factory->run();
