<?php

declare(strict_types=1);

namespace Temporal\Tests\Workflow;

use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;

#[Workflow\WorkflowInterface]
class SimpleSignaledWorkflow
{
    private $counter = 0;

    #[Workflow\SignalMethod(name: "add")]
    public function add(
        int $value
    ) {
        $this->counter += $value;
    }

    #[WorkflowMethod(name: 'SimpleSignaledWorkflow')]
    public function handler(): iterable
    {
        // collect signals during one second
        yield Workflow::timer(10);

        return $this->counter;
    }
}
