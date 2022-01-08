<?php

declare(strict_types=1);

namespace Arku\Newrelic\Transactions;

use Throwable;

final class TransactionDetail implements TransactionDetailInterface
{
    private array $data;
    private array $segments = [];
    private ?Throwable $throwable = null;

    public function __construct(array $data = [])
    {
        $this->data = $data;
    }

    public function setName(string $name): self
    {
        $this->data['transaction_name'] = $name;
        return $this;
    }

    public function setCustomData(string $key, string $value): self
    {
        $this->data[$key] = $value;
        return $this;
    }

    /**
     * @return string[]
     */
    public function getDetailsAsArray(): array
    {
        return $this->data;
    }

    public function addSegment(SegmentInterface $segment): self
    {
        $this->segments[] = $segment;
        return $this;
    }

    /**
     * @return array
     */
    public function getSegments(): array
    {
        return $this->segments;
    }

    public function setThrowable(Throwable $throwable): self
    {
        $this->throwable = $throwable;
        return $this;
    }

    public function getThrowable(): ?Throwable
    {
        return $this->throwable;
    }

}