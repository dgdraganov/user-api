using System;
using System.Text;
using RabbitMQ.Client;
using RabbitMQ.Client.Events;

class Program
{
    public static void Main()
    {
        Console.WriteLine("Starting RabbitMQ Consumer...");

        var rabbitHost = Environment.GetEnvironmentVariable("RABBITMQ_HOST") ?? "localhost";
        var rabbitUser = Environment.GetEnvironmentVariable("RABBITMQ_USER") ?? "guest";
        var rabbitPassword = Environment.GetEnvironmentVariable("RABBITMQ_PASSWORD") ?? "th15_15_s3cr3t";

        var factory = new ConnectionFactory()
        {
            HostName = rabbitHost,
            UserName = rabbitUser,
            Password = rabbitPassword,
            Port = 5672
        };

        using (var connection = factory.CreateConnection())
        using (var channel = connection.CreateModel())
        {
            channel.ExchangeDeclare(exchange: "users",
                                    type: ExchangeType.Topic,
                                    durable: true,
                                    autoDelete: false,
                                    arguments: null);

            var queueName = channel.QueueDeclare().QueueName;

            channel.QueueBind(queue: queueName,
                            exchange: "users",
                            routingKey: "user.event.*");

            Console.WriteLine("\tWaiting for messages... ");
            Console.WriteLine();

            var consumer = new EventingBasicConsumer(channel);
            consumer.Received += (model, ea) =>
            {
                var body = ea.Body.ToArray();
                var message = Encoding.UTF8.GetString(body);
                var routingKey = ea.RoutingKey;

                Console.WriteLine($"Received '{routingKey}': '{message}'");
            };

            channel.BasicConsume(queue: queueName,
                                autoAck: true,
                                consumer: consumer);

            Console.WriteLine("\tPress Ctrl+C to exit.");

            var exitEvent = new ManualResetEvent(false);
            Console.CancelKeyPress += (sender, eventArgs) =>
            {
                eventArgs.Cancel = true;
                exitEvent.Set();
            };
            exitEvent.WaitOne();
        }
    }
}