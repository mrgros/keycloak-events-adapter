#!/usr/bin/env tarantool

box.cfg {
    listen = 3301
}

queue = require('queue')
box.once("create_queue", function()
    queue.create_tube('events', 'fifottl', { if_not_exists = true })
    queue.create_tube('admin_events', 'fifottl', { if_not_exists = true })
end)
