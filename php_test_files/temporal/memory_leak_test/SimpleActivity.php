<?php

namespace Temporal\Tests\Activity;

use Temporal\Activity\ActivityInterface;
use Temporal\Activity\ActivityMethod;
use Temporal\Api\Common\V1\WorkflowExecution;
use Temporal\DataConverter\Bytes;
use Temporal\Tests\DTO\Message;
use Temporal\Tests\DTO\User;

require_once __DIR__ . '/SimpleWorkflow.php';

#[ActivityInterface(prefix: "SimpleActivity.")]
class SimpleActivity
{
    #[ActivityMethod]
    public function execute(): bool
    {
        return (bool)rand(0, 1);
    }
}