# Solana Monitor
Monitors Openbook Market Id creations and Raydium Liquidity Pool creations right from the blockchain.

### Discord Hook
Logs information in the configured Discord channels.

### Telegram Hook
Logs information in the configured Telegram chat. You can get the chat id for telegram by sending a message to the bot and going to ``https://api.telegram.org/bot<BOT_TOKEN>/getUpdates``, then look at message.chat.id within the result array.

### Custom Hooks
Custom hooks as well as altered hooks, can be requested with the developer of the bot against an additional fee.

### Addtional Notes
Some information for Raydium Liquidity Pools like the Embed Colour, Title Warning, and Openbook Costs are only available upon discovery of the Openbook Market Id creation. When just starting the bot some of this information is unavailable, because the Market Id was created prior to launching the bot. Usually after several minutes the bot is fully up to date and has all the Openbook information it requires.
