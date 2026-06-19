package com.learn.exceptions.dto;

public record TransferRequest(
    Long fromUserId,
    Long toUserId,
    double amount
) {}
