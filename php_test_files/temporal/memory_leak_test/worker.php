<?php

declare(strict_types=1);

require_once __DIR__ . '/../../vendor/autoload.php';

require_once __DIR__ . '/SimpleActivity.php';
require_once __DIR__ . '/SimpleWorkflow.php';

/**
 * @param string $dir
 * @return array<string>
 */
$getClasses = static function (string $dir): iterable {
    $files = glob($dir . '/*.php');

    foreach ($files as $file) {
        yield substr(basename($file), 0, -4);
    }
};


$factory = \Temporal\WorkerFactory::create();
$worker = $factory->newWorker('default');

$worker->registerWorkflowTypes(\Temporal\Tests\Workflow\SimpleWorkflow::class);
$worker->registerActivityImplementations(new \Temporal\Tests\Activity\SimpleActivity());

$factory->run();
