<?php

namespace Temporal\Tests\WorkflowWithLocalActivity;

use Temporal\Activity\LocalActivityOptions;
use Temporal\Tests\Activity\SimpleLocalActivity;
use Temporal\Workflow;
use Temporal\Workflow\SignalMethod;
use Temporal\Workflow\WorkflowMethod;

#[Workflow\WorkflowInterface]
class LoopWithSignalCoroutinesWorkflow
{
    private array $values = [];
    private array $result = [];
    private $simple;

    public function __construct()
    {
        $this->simple = Workflow::newActivityStub(
            SimpleLocalActivity::class,
            LocalActivityOptions::new()->withStartToCloseTimeout(5)
        );
    }

    #[SignalMethod]
    public function addValue(
        string $value
    ) {
        $value = yield $this->simple->prefix('in signal ', $value);
        $value = yield $this->simple->prefix('in signal 2 ', $value);

        $this->values[] = $value;
    }

    #[WorkflowMethod(name: 'LoopWithSignalCoroutinesWorkflow')]
    public function run(
        int $count
    ) {
        while (true) {
            yield Workflow::await(fn() => $this->values !== []);
            $value = array_shift($this->values);

            // uppercases
            $this->result[] = yield $this->simple->echo($value);

            if (count($this->result) === $count) {
                break;
            }
        }

        asort($this->result);
        return array_values($this->result);
    }
}
