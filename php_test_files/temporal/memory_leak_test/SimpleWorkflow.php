<?php

declare(strict_types=1);

namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityOptions;
use Temporal\Common\RetryOptions;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;
use Temporal\Tests\Activity\SimpleActivity;


require_once __DIR__ . '/SimpleActivity.php';

#[Workflow\WorkflowInterface]
class SimpleWorkflow
{
    private $activity;
    public function __construct()
    {
        $this->activity = Workflow::newActivityStub(
            SimpleActivity::class,
            ActivityOptions::new()
                ->withStartToCloseTimeout(5)
                ->withRetryOptions(
                    RetryOptions::new()->withMaximumAttempts(2)
                )
        );
    }

    #[WorkflowMethod(name: 'SimpleWorkflow')]
    public function handler(): iterable {
        while(true) {
            $result = yield $this->activity->execute();
                if ($result) {
                    break;
                 }
        }
        yield Workflow::timer(\Carbon\CarbonInterval::minutes(2));
    }
}
