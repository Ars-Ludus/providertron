# providertron
protron for short. Maybe vidertron so it doesn't sound so much like proton. Probably doesn't matter.
Data driven AI model provider written in Go, designed to be selectively efficient and self-healing.
It's basically just an alternative to using SDKs, but bundles it all together into an easy to manage connection between lib and app via JSON.
Feel free to modify it to add gRPC to adapt to Zig or RUST. I don't code in those, so I have no interest in setting up a test for that.
Also, Gemini will likely be my primary target for updates, since this is a tool to use in my own projects, so if a feature doesn't work with another provide, contribute a solution or submit an issue.
Contributions MUST use log/slog and properly wrap errors. Packages must handle their own errors.
