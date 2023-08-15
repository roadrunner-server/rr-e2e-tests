<?php

declare(strict_types=1);

namespace Temporal\Tests\Workflow;

use Carbon\CarbonInterval;
use Temporal\Activity\ActivityOptions;
use Temporal\Internal\Workflow\ActivityProxy;
use Temporal\Workflow;
use Temporal\Workflow\SignalMethod;

/**
 * Demonstrates asynchronous signalling of a workflow. Requires a local instance of Temporal server
 * to be running.
 */
#[Workflow\WorkflowInterface]
class SimpleWorkflow
{
    private array $input = [];
    private int $waitingNames = 0;
    private array $replay = [];
    private bool $exit = false;

    private ActivityProxy|SimpleActivity $greetingsActivity;

    public function __construct()
    {
        $this->greetingsActivity = Workflow::newActivityStub(
            SimpleActivity::class,
            ActivityOptions::new()
                ->withStartToCloseTimeout(CarbonInterval::seconds(20))
        );
    }

    #[Workflow\WorkflowMethod(name: 'SimpleWorkflow')]
    public function handler()
    {
        $result = [];
        while (true) {
            yield Workflow::await(fn() => $this->input !== [] || $this->exit);
            if ($this->input === [] && $this->exit) {
                // Wait signals
                if ($this->waitingNames !== 0) {
                    yield Workflow::await(fn() => $this->waitingNames === 0);
                    continue;
                }

                return $result;
            }

            $name = array_shift($this->input);
            $replay = array_shift($this->replay);
            $result[] = $name;
        }
    }

    #[SignalMethod]
    public function addName(string $name)
    {
        try {
            ++$this->waitingNames;
            $this->input[] = yield $this->greetingsActivity->composeStrings('Hello', $name);
            $this->replay[] = Workflow::isReplaying();
        } finally {
            --$this->waitingNames;
        }
    }

    #[SignalMethod]
    public function exit(): void
    {
        $this->exit = true;
    }
}