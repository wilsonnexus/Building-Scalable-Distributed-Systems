resource "aws_appautoscaling_target" "worker" {
  count              = var.enable_worker_autoscaling ? 1 : 0
  max_capacity       = 8
  min_capacity       = 1
  resource_id        = "service/${aws_ecs_cluster.main.name}/${aws_ecs_service.worker.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "worker_scale_out" {
  count              = var.enable_worker_autoscaling ? 1 : 0
  name               = "${var.project_name}-worker-scale-out"
  policy_type        = "StepScaling"
  resource_id        = aws_appautoscaling_target.worker[0].resource_id
  scalable_dimension = aws_appautoscaling_target.worker[0].scalable_dimension
  service_namespace  = aws_appautoscaling_target.worker[0].service_namespace

  step_scaling_policy_configuration {
    adjustment_type = "ChangeInCapacity"
    cooldown        = 60

    step_adjustment {
      metric_interval_lower_bound = 0
      scaling_adjustment          = 1
    }
  }
}

resource "aws_cloudwatch_metric_alarm" "queue_high" {
  count               = var.enable_worker_autoscaling ? 1 : 0
  alarm_name          = "${var.project_name}-queue-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ApproximateNumberOfMessagesVisible"
  namespace           = "AWS/SQS"
  period              = 60
  statistic           = "Average"
  threshold           = 20

  dimensions = {
    QueueName = aws_sqs_queue.jobs.name
  }

  alarm_actions = [aws_appautoscaling_policy.worker_scale_out[0].arn]
}