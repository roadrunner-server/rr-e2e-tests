<?php

namespace Temporal\Tests\WorkflowWithLocalActivity;

use Temporal\Activity\LocalActivityOptions;
use Temporal\DataConverter\Bytes;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;

#[Workflow\WorkflowInterface]
class BinaryWorkflow
{
    #[WorkflowMethod(name: 'BinaryWorkflow')]
    public function handler(
        Bytes $input
    ): iterable {
        $opts = LocalActivityOptions::new()->withStartToCloseTimeout(5);

        return yield Workflow::executeActivity('SimpleLocalActivity.sha512', [$input], $opts);
    }
}
