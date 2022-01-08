<?php

declare(strict_types=1);

namespace Arku\Tests\Newrelic;

use Arku\Newrelic\Response\EnrichResponse;
use Arku\Newrelic\Response\EnrichResponseInterface;
use Arku\Newrelic\Transactions\TransactionDetail;
use Arku\Newrelic\Transformers\TransactionDetailTransformer;
use Arku\Newrelic\Transformers\TransactionDetailTransformerInterface;
use Nyholm\Psr7\Response;
use PHPUnit\Framework\TestCase;

final class ExceptionTest extends TestCase
{
    public function testSimpleException()
    {
        $response = new Response();
        $transactionDetail = new TransactionDetail();
        $transactionDetail->setName('test');
        $transactionDetail->setThrowable(new \InvalidArgumentException('testMessage'));

        $enricher = new EnrichResponse(new TransactionDetailTransformer());
        $response = $enricher->enrich($response, $transactionDetail);
        $headers = $response->getHeaders();

        $this->assertArrayNotHasKey('rr_newrelic', $headers);

        $this->assertArrayHasKey(EnrichResponseInterface::ERROR_RR_REPORTING, $headers);
        $this->assertEquals(
            TransactionDetailTransformerInterface::THROWABLE_MESSAGE . ':testMessage',
            $headers[EnrichResponseInterface::ERROR_RR_REPORTING][0]
        );
        $this->assertEquals(
            TransactionDetailTransformerInterface::THROWABLE_CLASS . ':' . __FILE__,
            $headers[EnrichResponseInterface::ERROR_RR_REPORTING][1]
        );

        $this->assertEquals(
            strpos(
                TransactionDetailTransformerInterface::THROWABLE_STACKTRACE . ':',
                $headers[EnrichResponseInterface::ERROR_RR_REPORTING][2]
            ),
            0
        );
    }
}