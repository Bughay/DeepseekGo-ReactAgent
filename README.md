# DeepseekGo-ReactAgent
    // 1. Initialize memory with system prompt + tools + user prompt
    
    // 2. Set max iterations (e.g., 10)
    
    // 3. Loop:
    //    - Call a.oneloop()
    //    - Parse resp.Act:
    //        - if strings.HasPrefix(resp.Act, "finish|"): extract answer after "|", return
    //        - else: split by "|" to get toolName and args
    //    - Find tool in a.Tools by name
    //    - Execute tool with args, get result
    //    - Append exchange to a.Memory (assistant message with reasoning+act, user message with observation)
    //    - Update a.UserPrompt with observation for next iteration
    
    // 4. If max iterations reached, return error