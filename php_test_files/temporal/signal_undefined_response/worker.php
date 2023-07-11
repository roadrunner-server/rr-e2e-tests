<?php

declare(strict_types=1);

namespace Temporal\Tests\Workflow;

require_once __DIR__ . '/vendor/autoload.php';
require_once __DIR__ . '/SimpleActivity.php';
require_once __DIR__ . '/SimpleWorkflow.php';

$factory = \Temporal\WorkerFactory::create();
$worker = $factory->newWorker('default');

$worker->registerWorkflowTypes(SimpleWorkflow::class);
$worker->registerActivityImplementations(new SimpleActivity());

$factory->run();
