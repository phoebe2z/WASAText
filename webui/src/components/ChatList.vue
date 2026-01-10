<script>
export default {
    props: ['conversations', 'activeId'],
    emits: ['select-chat', 'create-group', 'create-dm'],
    data() {
        return {
            searchQuery: "",
            viewMode: 'list', // 'list', 'new_chat', 'new_contact'
            newContactName: "",
            errorMessage: ""
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
    },
    methods: {
        openNewChat() {
            this.viewMode = 'new_chat';
            this.searchQuery = "";
            this.errorMessage = "";
        },
        openNewContact() {
            this.viewMode = 'new_contact';
            this.errorMessage = "";
            this.newContactName = "";
        },
        goBack() {
             this.errorMessage = "";
             if (this.viewMode === 'new_contact') this.viewMode = 'new_chat';
             else this.viewMode = 'list';
        },
        async createDM() {
            if (!this.newContactName) return;
            this.errorMessage = "";
            
            try {
                let res = await this.$axios.post("/conversations", { recipientName: this.newContactName });
                // Success
                this.newContactName = "";
                this.viewMode = 'list';
                this.$emit('chat-created', res.data.conversationId);
            } catch (e) {
                if (e.response && e.response.status === 404) {
                    this.errorMessage = "User not found";
                } else if (e.response && e.response.status === 400) {
                     // Could be self-chat or bad request
                     this.errorMessage = "Invalid user (cannot chat with yourself)";
                } else {
                    this.errorMessage = "Error: " + (e.response ? e.response.statusText : e.message);
                }
            }
        },
        triggerCreateGroup() {
            this.$emit('create-group');
        }
    }
}
</script>

<template>
    <div class="d-flex flex-column h-100 bg-dark-list border-end border-secondary">
        
        <!-- VIEW: LIST -->
        <div v-if="viewMode === 'list'" class="d-flex flex-column h-100">
            <!-- Header -->
            <div class="p-3 bg-dark-header d-flex justify-content-between align-items-center">
                <h5 class="m-0 text-white">Chats</h5>
                <div class="d-flex gap-2">
                    <button class="btn btn-sm btn-link text-secondary p-0" title="New Chat" @click="openNewChat">
                        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-plus-square"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect><line x1="12" y1="8" x2="12" y2="16"></line><line x1="8" y1="12" x2="16" y2="12"></line></svg>
                    </button>
                </div>
            </div>

            <!-- List Search -->
            <div class="p-2 border-bottom border-secondary">
                <input type="text" class="form-control form-control-sm bg-dark-input text-white border-0" placeholder="Search or start new chat" v-model="searchQuery">
            </div>

            <!-- Conversations List -->
            <div class="flex-grow-1 overflow-auto custom-scrollbar">
                <div 
                    v-for="c in filteredConversations" 
                    :key="c.conversationId"
                    class="d-flex align-items-center p-3 border-bottom border-secondary chat-item"
                    :class="{ 'active-chat': activeId === c.conversationId }"
                    @click="$emit('select-chat', c.conversationId)"
                >
                    <!-- Avatar -->
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

        <!-- VIEW: NEW CHAT -->
        <div v-else-if="viewMode === 'new_chat'" class="d-flex flex-column h-100">
             <!-- Header -->
            <div class="p-3 bg-dark-header d-flex align-items-center gap-3">
                 <button class="btn btn-link text-white p-0" @click="goBack">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-arrow-left"><line x1="19" y1="12" x2="5" y2="12"></line><polyline points="12 19 5 12 12 5"></polyline></svg>
                 </button>
                <h5 class="m-0 text-white">New chat</h5>
            </div>
            
            <div class="p-2">
                 <!-- Search (Visual only for now or contact search?) -->
                <input type="text" class="form-control form-control-sm bg-dark-input text-white border-0 mb-3" placeholder="Search name or number">
                
                <!-- Menu Items -->
                <div class="d-flex align-items-center p-3 text-white chat-item rounded" @click="triggerCreateGroup">
                     <div class="rounded-circle bg-success d-flex justify-content-center align-items-center me-3" style="width: 45px; height: 45px;">
                         <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-users"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg>
                     </div>
                     <span>New group</span>
                </div>
                
                <div class="d-flex align-items-center p-3 text-white chat-item rounded" @click="openNewContact">
                     <div class="rounded-circle bg-success d-flex justify-content-center align-items-center me-3" style="width: 45px; height: 45px;">
                         <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-user-plus"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="8.5" cy="7" r="4"></circle><line x1="20" y1="8" x2="20" y2="14"></line><line x1="23" y1="11" x2="17" y2="11"></line></svg>
                     </div>
                     <span>New contact</span>
                </div>
            </div>
        </div>

        <!-- VIEW: NEW CONTACT -->
         <div v-else-if="viewMode === 'new_contact'" class="d-flex flex-column h-100">
             <!-- Header -->
            <div class="p-3 bg-dark-header d-flex align-items-center gap-3">
                 <button class="btn btn-link text-white p-0" @click="goBack">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-arrow-left"><line x1="19" y1="12" x2="5" y2="12"></line><polyline points="12 19 5 12 12 5"></polyline></svg>
                 </button>
                <h5 class="m-0 text-white">New contact</h5>
            </div>
            
            <div class="p-4">
                <div class="mb-3">
                    <label class="form-label text-secondary small">Username</label>
                    <input 
                        v-model="newContactName" 
                        @input="errorMessage = ''"
                        type="text" 
                        class="form-control bg-dark-input text-white border-secondary" 
                        placeholder="Enter Username"
                    >
                    <div v-if="errorMessage" class="text-danger small mt-1">{{ errorMessage }}</div>
                </div>
                <button class="btn btn-success w-100" @click="createDM" :disabled="!newContactName">Start Chat</button>
            </div>
        </div>

    </div>
</template>

<style scoped>
.bg-dark-list { background-color: #111b21; }
.bg-dark-header { background-color: #202c33; }
.bg-dark-input { background-color: #202c33; }
.bg-dark-input:focus { background-color: #202c33; color: white; border-color: #00a884; box-shadow: none; }

.chat-item { cursor: pointer; transition: background-color 0.2s; }
.chat-item:hover { background-color: #202c33; }
.active-chat { background-color: #2a3942; }

.text-secondary { color: #8696a0 !important; }
.bg-success { background-color: #00a884 !important; } /* WhatsApp Green */

/* Custom Scrollbar for Webkit */
.custom-scrollbar::-webkit-scrollbar { width: 6px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background-color: #374045; border-radius: 3px; }
</style>
