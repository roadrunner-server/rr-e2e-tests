<?php
# Generated by the protocol buffer compiler (roadrunner-server/grpc). DO NOT EDIT!
# source: test.proto

namespace Test;

use Spiral\RoadRunner\GRPC;

interface TestInterface extends GRPC\ServiceInterface
{
    // GRPC specific service name.
    public const NAME = "test.Test";

    /**
    * @param GRPC\ContextInterface $ctx
    * @param Message $in
    * @return Message
    *
    * @throws GRPC\Exception\InvokeException
    */
    public function Echo(GRPC\ContextInterface $ctx, Message $in): Message;

    /**
    * @param GRPC\ContextInterface $ctx
    * @param Message $in
    * @return Message
    *
    * @throws GRPC\Exception\InvokeException
    */
    public function Throw(GRPC\ContextInterface $ctx, Message $in): Message;

    /**
    * @param GRPC\ContextInterface $ctx
    * @param Message $in
    * @return Message
    *
    * @throws GRPC\Exception\InvokeException
    */
    public function Die(GRPC\ContextInterface $ctx, Message $in): Message;

    /**
    * @param GRPC\ContextInterface $ctx
    * @param Message $in
    * @return Message
    *
    * @throws GRPC\Exception\InvokeException
    */
    public function Info(GRPC\ContextInterface $ctx, Message $in): Message;

    /**
    * @param GRPC\ContextInterface $ctx
    * @param EmptyMessage $in
    * @return EmptyMessage
    *
    * @throws GRPC\Exception\InvokeException
    */
    public function Ping(GRPC\ContextInterface $ctx, EmptyMessage $in): EmptyMessage;
}
