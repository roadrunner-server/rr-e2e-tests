<?php

declare(strict_types=1);

namespace Arku\Newrelic\Response;

use Arku\Newrelic\Transactions\TransactionDetailInterface;
use Arku\Newrelic\Transformers\TransactionDetailTransformerInterface;
use Psr\Http\Message\ResponseInterface;

final class EnrichResponse implements EnrichResponseInterface
{
    private const SEGMENT_PATTERN = 'nr_segment%d';

    private TransactionDetailTransformerInterface $transformer;

    public function __construct(TransactionDetailTransformerInterface $transformer)
    {
        $this->transformer = $transformer;
    }

    public function enrich(ResponseInterface $response, TransactionDetailInterface $transactionDetail): ResponseInterface
    {
        if ($throwable = $transactionDetail->getThrowable()) {
            $throwableData = $this->transformer->transformThrowable($throwable);
            return $response->withHeader(self::ERROR_RR_REPORTING, $throwableData);
        }

        $response =  $response->withHeader(self::MAIN_RR_HEADER, $this->transformer->transform($transactionDetail));

        if ($segments = $transactionDetail->getSegments()) {
            $keys = $this->generateKeys(count($segments));
            $segments = array_combine($keys, $segments);
            $response = $response->withHeader(self::SEGMENTS_RR_HEADER, implode(',', $keys));
            foreach ($segments as $key => $segment) {
                $response = $response->withHeader($key, $this->transformer->transformSegment($segment));
            }
        }

        return $response;
    }

    private function generateKeys(int $count): array
    {
        $segments = [];
        for ($i=0;$i< $count; $i++) {
            $segments[] = sprintf(self::SEGMENT_PATTERN, $count);
        }
        return $segments;
    }
}