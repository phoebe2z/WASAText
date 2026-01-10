<script>
export default {
    props: ['conversations', 'activeId'],
    emits: ['select-chat', 'create-group'],
    data() {
        return {
            searchQuery: ""
        }
    },
    computed: {
        filteredConversations() {
            if (!this.searchQuery) return this.conversations;
            const query = this.searchQuery.toLowerCase();
            return this.conversations.filter(c => 
                (c.name && c.name.toLowerCase().includes(query)) ||
                (c.latestMessagePreview && c.latestMessagePreview.toLowerCase().includes(query))
            );
        }
    }
}
</script>

<template>
    <div class="d-flex flex-column h-100 bg-dark-list border-end border-secondary">
        <!-- Header -->
        <div class="p-3 bg-dark-header d-flex justify-content-between align-items-center">
            <h5 class="m-0 text-white">Chats</h5>
            <div class="d-flex gap-2">
                <button class="btn btn-sm btn-link text-secondary p-0" title="New Chat" @click="$emit('create-group')">
                    <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-plus-square"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect><line x1="12" y1="8" x2="12" y2="16"></line><line x1="8" y1="12" x2="16" y2="12"></line></svg>
                </button>
                <button class="btn btn-sm btn-link text-secondary p-0" title="Filter" disabled>
                    <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-filter"><polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3"></polygon></svg>
                </button>
            </div>
        </div>

        <!-- Search -->
        <div class="p-2 border-bottom border-secondary">
            <input type="text" class="form-control form-control-sm bg-dark-input text-white border-0" placeholder="Search or start new chat" v-model="searchQuery">
        </div>

        <!-- List -->
        <div class="flex-grow-1 overflow-auto custom-scrollbar">
            <div 
                v-for="c in filteredConversations" 
                :key="c.conversationId"
                class="d-flex align-items-center p-3 border-bottom border-secondary chat-item"
                :class="{ 'active-chat': activeId === c.conversationId }"
                @click="$emit('select-chat', c.conversationId)"
            >
                <!-- Avatar Placeholder -->
                <div class="rounded-circle bg-secondary d-flex justify-content-center align-items-center me-3 text-white" style="width: 45px; height: 45px; flex-shrink: 0;">
                    <span v-if="c.photoUrl" class="w-100 h-100 rounded-circle overflow-hidden">
                         <img :src="c.photoUrl" class="w-100 h-100" style="object-fit: cover;" alt="C">
                    </span>
                    <span v-else>{{ (c.name || 'C').charAt(0).toUpperCase() }}</span>
                </div>
                
                <div class="flex-grow-1 min-w-0">
                    <div class="d-flex justify-content-between align-items-baseline mb-1">
                        <h6 class="mb-0 text-white text-truncate">{{ c.name || (c.isGroup ? 'Group ' + c.conversationId : 'Chat ' + c.conversationId) }}</h6>
                        <small class="text-secondary" style="font-size: 0.75rem;">{{ c.lastMessageTime ? new Date(c.lastMessageTime).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'}) : '' }}</small>
                    </div>
                    <div class="d-flex justify-content-between align-items-center">
                        <small class="text-secondary text-truncate d-block" style="max-width: 90%;">{{ c.latestMessagePreview }}</small>
                        <span v-if="c.unreadCount > 0" class="badge rounded-pill bg-success">{{ c.unreadCount }}</span>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.bg-dark-list { background-color: #111b21; }
.bg-dark-header { background-color: #202c33; }
.bg-dark-input { background-color: #202c33; }

.chat-item { cursor: pointer; transition: background-color 0.2s; }
.chat-item:hover { background-color: #202c33; }
.active-chat { background-color: #2a3942; }

.text-secondary { color: #8696a0 !important; }

/* Custom Scrollbar for Webkit */
.custom-scrollbar::-webkit-scrollbar { width: 6px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background-color: #374045; border-radius: 3px; }
</style>
