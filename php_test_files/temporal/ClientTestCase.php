<?php

require __DIR__ . '/../vendor/autoload.php';

use Temporal\Api\Enums\V1\EventType;
use Temporal\Api\History\V1\HistoryEvent;
use Temporal\Api\Workflowservice\V1\GetWorkflowExecutionHistoryRequest;
use Temporal\Client\GRPC\ServiceClient;
use Temporal\Client\WorkflowClient;
use Temporal\Tests\Functional\FunctionalTestCase;
use Temporal\Workflow\WorkflowExecution;

$client = new WorkflowClient(ServiceClient::create('localhost:7233'));

$simple = $client->newWorkflowStub(\Temporal\Tests\Workflow\SimpleWorkflow::class);
$simple->handler('hello world');
