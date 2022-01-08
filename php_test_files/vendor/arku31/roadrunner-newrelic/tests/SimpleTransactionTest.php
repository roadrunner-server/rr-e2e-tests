<?php

declare(strict_types=1);

namespace Arku\Tests\Newrelic;

use Arku\Newrelic\Response\EnrichResponse;
use Arku\Newrelic\Transactions\Segment;
use Arku\Newrelic\Transactions\TransactionDetail;
use Arku\Newrelic\Transformers\TransactionDetailTransformer;
use Nyholm\Psr7\Response;
use PHPUnit\Framework\TestCase;

final class SimpleTransactionTest extends TestCase
{
    public function testMvp()
    {
        $response = new Response();
        $transactionDetail = new TransactionDetail();
        $transactionDetail->setName('test');
        $transactionDetail->setCustomData('key', 'value');

        $segment = new Segment();
        $segment->setName('testSegment');
        $segment->setDuration('1');
        $segment->setMeta(['testmetakey' => 'testmetavalue']);

        $transactionDetail->addSegment($segment);

        $enricher = new EnrichResponse(new TransactionDetailTransformer());
        $response = $enricher->enrich($response, $transactionDetail);
        $headers = $response->getHeaders();

        $this->assertArrayHasKey('rr_newrelic', $headers);
        $this->assertEquals('transaction_name:test', $headers['rr_newrelic'][0]);
        $this->assertEquals('key:value', $headers['rr_newrelic'][1]);

        $this->assertArrayHasKey('rr_newrelic_headers', $headers);
        $this->assertEquals('nr_segment1', $headers['rr_newrelic_headers'][0]);

        $this->assertArrayHasKey('nr_segment1', $headers);
        $this->assertEquals([
            'name:testSegment',
            'duration:1000',
            'testmetakey:testmetavalue'
        ], $headers['nr_segment1']);
    }
}