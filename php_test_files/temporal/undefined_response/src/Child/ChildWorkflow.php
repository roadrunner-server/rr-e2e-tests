<?php

/**
 * This file is part of Temporal package.
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

declare(strict_types=1);

namespace Temporal\Samples\Child;

use Carbon\CarbonInterval;
use Temporal\Activity\ActivityOptions;
use Temporal\Common\RetryOptions;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowInterface;
use Temporal\Workflow\WorkflowMethod;

#[WorkflowInterface]
class ChildWorkflow
{
    private $greetingActivity;

    public function __construct()
    {
        /**
         * To enable activity retry set {@link RetryOptions} on {@link ActivityOptions}.
         */
        $this->greetingActivity = Workflow::newActivityStub(
            GreetingActivity::class,
            ActivityOptions::new()
                ->withScheduleToCloseTimeout(CarbonInterval::seconds(20))
                ->withRetryOptions(
                    RetryOptions::new()
                        ->withInitialInterval(CarbonInterval::seconds(2))
                        ->withMaximumAttempts(2)
                        ->withNonRetryableExceptions([\InvalidArgumentException::class])
                )
        );
    }

    #[WorkflowMethod("Child.greet")]
    public function greet(
        string $name
    ) {
        \file_put_contents('runtime/TIMING_CHILD.txt', \date('r') . " before activity\n", FILE_APPEND);
        yield $this->greetingActivity->composeGreeting('Hello', $name);
        \file_put_contents('runtime/TIMING_CHILD.txt', \date('r') . " after activity\n", FILE_APPEND);

        return 'Hello ' . $name . ' from child workflow!';
    }
}