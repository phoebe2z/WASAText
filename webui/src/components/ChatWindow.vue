<script>
export default {
    props: ['conversation', 'messages', 'currentUser', 'userId'],
    emits: ['send-message', 'delete-message', 'toggle-info', 'react-message', 'forward-message', 'unreact-message'],
    data() {
        return {
            newMessage: "",
            commonEmojis: ["👍", "❤️", "😂", "😮", "😢", "😡"],
            showReactionFor: null,
            showEmojiPicker: false
        }
    },
    methods: {
        sendMessage() {
            if (!this.newMessage.trim()) return;
            this.$emit('send-message', this.newMessage);
            this.newMessage = "";
            this.showEmojiPicker = false;
        },
        addEmoji(emoji) {
            this.newMessage += emoji;
            this.showEmojiPicker = false;
        },
        toggleReaction(msgId) {
             if (this.showReactionFor === msgId) {
                 this.showReactionFor = null;
             } else {
                 this.showReactionFor = msgId;
             }
        },
        react(msgId, emoji) {
             this.$emit('react-message', msgId, emoji);
             this.showReactionFor = null;
        },
        startForward(msg) {
             this.$emit('forward-message', msg);
        },
        formatTime(t) {
            return new Date(t).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
        },
        scrollToBottom() {
            this.$nextTick(() => {
                const container = this.$refs.msgContainer;
                if (container) container.scrollTop = container.scrollHeight;
            });
        }
    },
    watch: {
        messages: {
            handler() {
                this.scrollToBottom();
            },
            deep: true
        }
    },
    mounted() {
        this.scrollToBottom();
    }
}
</script>

<template>
    <div class="d-flex flex-column h-100 bg-image">
        <!-- Header -->
        <div class="p-2 px-3 bg-dark-header border-start border-secondary d-flex align-items-center justify-content-between" style="cursor: pointer;" @click="$emit('toggle-info')">
            <div class="d-flex align-items-center">
                 <div class="rounded-circle bg-secondary d-flex justify-content-center align-items-center me-3 text-white overflow-hidden" style="width: 40px; height: 40px;">
                    <img v-if="conversation.photoUrl" :src="conversation.photoUrl" class="w-100 h-100" style="object-fit: cover;">
                    <span v-else>{{ (conversation.name || 'C').charAt(0).toUpperCase() }}</span>
                </div>
                <div>
                     <h6 class="mb-0 text-white">{{ conversation.name || (conversation.isGroup ? 'Group ' + conversation.conversationId : 'Chat ' + conversation.conversationId) }}</h6>
                     <small class="text-secondary" v-if="conversation.isGroup">click here for group info</small>
                </div>
            </div>
            <div class="d-flex gap-3 text-secondary">
                 <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-search"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
                 <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-more-vertical"><circle cx="12" cy="12" r="1"></circle><circle cx="12" cy="5" r="1"></circle><circle cx="12" cy="19" r="1"></circle></svg>
            </div>
        </div>

        <!-- Messages Area -->
        <div class="flex-grow-1 overflow-auto p-4 custom-scrollbar position-relative" ref="msgContainer">
             <div v-for="msg in messages" :key="msg.id" class="d-flex flex-column mb-3" :class="{ 'align-items-end': parseInt(msg.senderId) === parseInt(userId), 'align-items-start': parseInt(msg.senderId) !== parseInt(userId) }">
                <!-- Message Bubble -->
                <div class="message-bubble p-2 rounded-3 shadow-sm position-relative parent-hover-trigger" :class="{ 'message-out': parseInt(msg.senderId) === parseInt(userId), 'message-in': parseInt(msg.senderId) !== parseInt(userId) }">
                     <!-- Sender Name (if Group and not me) -->
                     <small v-if="conversation.isGroup && parseInt(msg.senderId) !== parseInt(userId)" class="text-warning fw-bold d-block mb-1" style="font-size: 0.75rem;">{{ msg.senderName || ('User ' + msg.senderId) }}</small>
                     
                     <div class="text-white message-text">{{ msg.content }}</div>
                     
                     <!-- Reactions Display -->
                     <div v-if="msg.reactions && msg.reactions.length > 0" class="d-flex flex-wrap gap-1 mt-1">
                         <span v-for="r in msg.reactions" :key="r.user_id + r.emoticon" class="badge bg-secondary rounded-pill" style="font-size: 0.7em; cursor: pointer;" title="Click to remove (if yours)" @click="$emit('unreact-message', msg.id, r.user_id)">
                             {{ r.emoticon }}
                         </span>
                     </div>

                     <div class="d-flex justify-content-end align-items-center gap-2 mt-1">
                         <span class="text-white-50 time-text">{{ formatTime(msg.timeStamp) }}</span>
                         
                         <!-- Actions Dropdown/Hover -->
                         <!-- Reaction Button -->
                         <div class="position-relative">
                             <button class="btn btn-link p-0 text-white-50 action-btn" @click.stop="toggleReaction(msg.id)">
                                 <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-smile"><circle cx="12" cy="12" r="10"></circle><path d="M8 14s1.5 2 4 2 4-2 4-2"></path><line x1="9" y1="9" x2="9.01" y2="9"></line><line x1="15" y1="9" x2="15.01" y2="9"></line></svg>
                             </button>
                             <!-- Reaction Popover -->
                             <div v-if="showReactionFor === msg.id" class="position-absolute bottom-100 start-50 translate-middle-x bg-dark border border-secondary rounded shadow p-1 d-flex gap-1" style="z-index: 1000; width: max-content;">
                                 <button v-for="emoji in commonEmojis" :key="emoji" class="btn btn-sm btn-link text-decoration-none p-1 fs-5" @click="react(msg.id, emoji)">{{ emoji }}</button>
                             </div>
                         </div>
                         
                         <!-- Forward -->
                         <button class="btn btn-link p-0 text-white-50 action-btn" @click="startForward(msg)">
                             <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-share"><path d="M4 12v8a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-8"></path><polyline points="16 6 12 2 8 6"></polyline><line x1="12" y1="2" x2="12" y2="15"></line></svg>
                         </button>

                         <!-- Delete -->
                         <button v-if="msg.senderName === currentUser" class="btn btn-link p-0 text-white-50 action-btn" @click="$emit('delete-message', msg.id)">
                             <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-trash-2"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>
                         </button>
                     </div>
                </div>
             </div>
        </div>

        <!-- Input Area -->
        <div class="p-2 bg-dark-header d-flex align-items-center gap-2">
             <!-- Emoji Picker Container -->
             <div class="position-relative">
                <button class="btn btn-link text-secondary p-2" @click="showEmojiPicker = !showEmojiPicker">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-smile"><circle cx="12" cy="12" r="10"></circle><path d="M8 14s1.5 2 4 2 4-2 4-2"></path><line x1="9" y1="9" x2="9.01" y2="9"></line><line x1="15" y1="9" x2="15.01" y2="9"></line></svg>
                </button>
                <!-- Emoji Picker Popover -->
                <div v-if="showEmojiPicker" class="position-absolute bottom-100 start-0 bg-dark border border-secondary rounded shadow p-2 d-flex flex-wrap gap-2" style="z-index: 1000; width: 180px;">
                    <button v-for="emoji in commonEmojis" :key="emoji" class="btn btn-sm btn-link text-decoration-none p-1 fs-4" @click="addEmoji(emoji)">{{ emoji }}</button>
                </div>
             </div>
             
             <input type="text" class="form-control bg-dark-input text-white border-0 rounded-3 py-2" placeholder="Type a message" v-model="newMessage" @keyup.enter="sendMessage">
             
             <button class="btn btn-link text-secondary p-2" @click="sendMessage" :disabled="!newMessage.trim()">
                  <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-send"><line x1="22" y1="2" x2="11" y2="13"></line><polygon points="22 2 15 22 11 13 2 9 22 2"></polygon></svg>
             </button>
        </div>
    </div>
</template>

<style scoped>
.bg-image {
    background-color: #0b141a;
    background-image: url("https://user-images.githubusercontent.com/15075759/28719144-86dc0f70-73b1-11e7-911d-60d70fcded21.png");
    background-repeat: repeat;
    background-size: 400px;
    /* Blend with dark background */
}
.bg-dark-header { background-color: #202c33; }
.bg-dark-input { background-color: #2a3942; }

.message-bubble {
    max-width: 65%;
    position: relative;
    font-size: 14.2px;
    line-height: 19px;
}
.message-in { background-color: #202c33; border-top-left-radius: 0 !important; }
.message-out { background-color: #005c4b; border-top-right-radius: 0 !important; margin-left: auto;}

.time-text { font-size: 11px; margin-top: 2px;}

.text-secondary { color: #8696a0 !important; }
.text-warning { color: #e5b955 !important; } /* Group sender name color */

/* Scrollbar */
.custom-scrollbar::-webkit-scrollbar { width: 6px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background-color: #374045; border-radius: 3px; }
</style>
