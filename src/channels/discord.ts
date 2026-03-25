import {
  Client,
  Events,
  GatewayIntentBits,
  Message,
  TextChannel,
} from 'discord.js';

import { ASSISTANT_NAME, TRIGGER_PATTERN } from '../config.js';
import { readEnvFile } from '../env.js';
import { logger } from '../logger.js';
import { registerChannel, ChannelOpts } from './registry.js';
import {
  Channel,
  OnChatMetadata,
  OnInboundMessage,
  RegisteredGroup,
} from '../types.js';

export interface DiscordChannelOpts {
  onMessage: OnInboundMessage;
  onChatMetadata: OnChatMetadata;
  registeredGroups: () => Record<string, RegisteredGroup>;
}

export class DiscordChannel implements Channel {
  name: string;

  private client: Client | null = null;
  private opts: DiscordChannelOpts;
  private botToken: string;

  constructor(
    botToken: string,
    opts: DiscordChannelOpts,
    channelName = 'discord',
  ) {
    this.botToken = botToken;
    this.opts = opts;
    this.name = channelName;
  }

  async connect(): Promise<void> {
    this.client = new Client({
      intents: [
        GatewayIntentBits.Guilds,
        GatewayIntentBits.GuildMessages,
        GatewayIntentBits.MessageContent,
        GatewayIntentBits.DirectMessages,
      ],
    });

    this.client.on(Events.MessageCreate, async (message: Message) => {
      // Ignore bot messages (including own)
      if (message.author.bot) return;

      const channelId = message.channelId;
      const chatJid = `dc:${channelId}`;
      let content = message.content;
      const timestamp = message.createdAt.toISOString();
      const senderName =
        message.member?.displayName ||
        message.author.displayName ||
        message.author.username;
      const sender = message.author.id;
      const msgId = message.id;

      // Determine chat name
      let chatName: string;
      if (message.guild) {
        const textChannel = message.channel as TextChannel;
        chatName = `${message.guild.name} #${textChannel.name}`;
      } else {
        chatName = senderName;
      }

      // Translate Discord @bot mentions into TRIGGER_PATTERN format.
      // Discord mentions look like <@botUserId> — these won't match
      // TRIGGER_PATTERN (e.g., ^@Andy\b), so we prepend the trigger
      // when the bot is @mentioned.
      if (this.client?.user) {
        const botId = this.client.user.id;
        const group = this.opts.registeredGroups()[chatJid];
        const triggerName = group?.assistantName || ASSISTANT_NAME;

        const isBotMentioned =
          message.mentions.users.has(botId) ||
          content.includes(`<@${botId}>`) ||
          content.includes(`<@!${botId}>`);

        // Also detect role mentions (<@&roleId>) where the role name
        // matches the assistant name — Discord converts @pm to a role
        // mention when a role named "pm" exists.
        const isRoleMentionTrigger =
          !isBotMentioned &&
          message.mentions.roles?.some(
            (role) => role.name.toLowerCase() === triggerName.toLowerCase(),
          );

        if (isBotMentioned || isRoleMentionTrigger) {
          if (isBotMentioned) {
            // Strip the <@botId> mention to avoid visual clutter
            content = content
              .replace(new RegExp(`<@!?${botId}>`, 'g'), '')
              .trim();
          }
          if (isRoleMentionTrigger) {
            // Strip the role mention(s) that matched the trigger
            message.mentions.roles.forEach((role) => {
              if (role.name.toLowerCase() === triggerName.toLowerCase()) {
                content = content
                  .replace(new RegExp(`<@&${role.id}>`, 'g'), '')
                  .trim();
              }
            });
          }
          const triggerRe = new RegExp(
            `^@${triggerName.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}\\b`,
            'i',
          );
          // Prepend trigger if not already present
          if (!triggerRe.test(content)) {
            content = `@${triggerName} ${content}`;
          }
        }
      }

      // Handle attachments — store placeholders so the agent knows something was sent
      if (message.attachments.size > 0) {
        const attachmentDescriptions = [...message.attachments.values()].map(
          (att) => {
            const contentType = att.contentType || '';
            if (contentType.startsWith('image/')) {
              return `[Image: ${att.name || 'image'}]`;
            } else if (contentType.startsWith('video/')) {
              return `[Video: ${att.name || 'video'}]`;
            } else if (contentType.startsWith('audio/')) {
              return `[Audio: ${att.name || 'audio'}]`;
            } else {
              return `[File: ${att.name || 'file'}]`;
            }
          },
        );
        if (content) {
          content = `${content}\n${attachmentDescriptions.join('\n')}`;
        } else {
          content = attachmentDescriptions.join('\n');
        }
      }

      // Handle reply context — include who the user is replying to
      if (message.reference?.messageId) {
        try {
          const repliedTo = await message.channel.messages.fetch(
            message.reference.messageId,
          );
          const replyAuthor =
            repliedTo.member?.displayName ||
            repliedTo.author.displayName ||
            repliedTo.author.username;
          content = `[Reply to ${replyAuthor}] ${content}`;
        } catch {
          // Referenced message may have been deleted
        }
      }

      // Store chat metadata for discovery
      const isGroup = message.guild !== null;
      this.opts.onChatMetadata(
        chatJid,
        timestamp,
        chatName,
        'discord',
        isGroup,
      );

      // Only deliver full message for registered groups owned by this bot.
      // When multiple Discord bots coexist, skip messages for groups
      // that belong to a different bot (based on channelName).
      const group = this.opts.registeredGroups()[chatJid];
      if (group?.channelName && group.channelName !== this.name) return;
      if (!group) {
        logger.debug(
          { chatJid, chatName },
          'Message from unregistered Discord channel',
        );
        return;
      }

      // Deliver message — startMessageLoop() will pick it up
      this.opts.onMessage(chatJid, {
        id: msgId,
        chat_jid: chatJid,
        sender,
        sender_name: senderName,
        content,
        timestamp,
        is_from_me: false,
      });

      logger.info(
        { chatJid, chatName, sender: senderName },
        'Discord message stored',
      );
    });

    // Handle errors gracefully
    this.client.on(Events.Error, (err) => {
      logger.error({ err: err.message }, 'Discord client error');
    });

    return new Promise<void>((resolve) => {
      this.client!.once(Events.ClientReady, (readyClient) => {
        logger.info(
          { username: readyClient.user.tag, id: readyClient.user.id },
          'Discord bot connected',
        );
        console.log(`\n  Discord bot: ${readyClient.user.tag}`);
        console.log(
          `  Use /chatid command or check channel IDs in Discord settings\n`,
        );
        resolve();
      });

      this.client!.login(this.botToken);
    });
  }

  async sendMessage(jid: string, text: string): Promise<void> {
    if (!this.client) {
      logger.warn('Discord client not initialized');
      return;
    }

    try {
      const channelId = jid.replace(/^dc:/, '');
      const channel = await this.client.channels.fetch(channelId);

      if (!channel || !('send' in channel)) {
        logger.warn({ jid }, 'Discord channel not found or not text-based');
        return;
      }

      const textChannel = channel as TextChannel;

      // Discord has a 2000 character limit per message — split if needed
      const MAX_LENGTH = 2000;
      if (text.length <= MAX_LENGTH) {
        await textChannel.send(text);
      } else {
        for (let i = 0; i < text.length; i += MAX_LENGTH) {
          await textChannel.send(text.slice(i, i + MAX_LENGTH));
        }
      }
      logger.info({ jid, length: text.length }, 'Discord message sent');
    } catch (err) {
      logger.error({ jid, err }, 'Failed to send Discord message');
    }
  }

  isConnected(): boolean {
    return this.client !== null && this.client.isReady();
  }

  ownsJid(jid: string): boolean {
    if (!jid.startsWith('dc:')) return false;
    // When multiple Discord bots coexist, check the registered group's channelName
    // to determine which bot instance owns this JID.
    const group = this.opts.registeredGroups()[jid];
    if (group?.channelName) {
      return group.channelName === this.name;
    }
    // Fallback: check channel cache for groups without explicit channelName
    if (this.client?.isReady()) {
      const channelId = jid.replace(/^dc:/, '');
      return this.client.channels.cache.has(channelId);
    }
    return true;
  }

  async disconnect(): Promise<void> {
    if (this.client) {
      this.client.destroy();
      this.client = null;
      logger.info('Discord bot stopped');
    }
  }

  async setTyping(jid: string, isTyping: boolean): Promise<void> {
    if (!this.client || !isTyping) return;
    try {
      const channelId = jid.replace(/^dc:/, '');
      const channel = await this.client.channels.fetch(channelId);
      if (channel && 'sendTyping' in channel) {
        await (channel as TextChannel).sendTyping();
      }
    } catch (err) {
      logger.debug({ jid, err }, 'Failed to send Discord typing indicator');
    }
  }
}

registerChannel('discord', (opts: ChannelOpts) => {
  const envVars = readEnvFile(['DISCORD_BOT_TOKEN']);
  const token =
    process.env.DISCORD_BOT_TOKEN || envVars.DISCORD_BOT_TOKEN || '';
  if (!token) {
    logger.warn('Discord: DISCORD_BOT_TOKEN not set');
    return null;
  }
  return new DiscordChannel(token, opts, 'discord');
});

registerChannel('discord2', (opts: ChannelOpts) => {
  const envVars = readEnvFile(['DISCORD_BOT_TOKEN_2']);
  const token =
    process.env.DISCORD_BOT_TOKEN_2 || envVars.DISCORD_BOT_TOKEN_2 || '';
  if (!token) {
    return null;
  }
  return new DiscordChannel(token, opts, 'discord2');
});

registerChannel('discord3', (opts: ChannelOpts) => {
  const envVars = readEnvFile(['DISCORD_BOT_TOKEN_3']);
  const token =
    process.env.DISCORD_BOT_TOKEN_3 || envVars.DISCORD_BOT_TOKEN_3 || '';
  if (!token) {
    return null;
  }
  return new DiscordChannel(token, opts, 'discord3');
});
