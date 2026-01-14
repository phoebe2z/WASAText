<script>
export default {
    props: ['conversations', 'activeId', 'currentUserId'],
    emits: ['select-chat', 'create-group', 'create-dm', 'show-error'],
    data() {
        return {
            searchQuery: "",
            viewMode: 'list', // 'list', 'new_chat', 'new_contact', 'new_group', 'group_details'
            newContactName: "",
            errorMessage: "",
            
            // New Group Flow
            allUsers: [],
            selectedUsers: [], // Array of user objects
            groupName: "",
            groupPhoto: null,
            groupSearch: ""
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
        },
        filteredUsers() {
            const query = this.groupSearch.toLowerCase();
            return this.allUsers.filter(u => 
                u.name.toLowerCase().includes(query) && 
                u.id.toString() !== this.currentUserId?.toString()
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
        async openNewGroup() {
            this.viewMode = 'new_group';
            this.selectedUsers = [];
            this.groupSearch = "";
            await this.fetchUsers();
        },
        async fetchUsers() {
            try {
                let res = await this.$axios.get("/users");
                this.allUsers = res.data;
            } catch (e) {
                console.error("Error fetching users", e);
            }
        },
        toggleUserSelection(user) {
            const idx = this.selectedUsers.findIndex(u => u.id === user.id);
            if (idx > -1) {
                this.selectedUsers.splice(idx, 1);
            } else {
                this.selectedUsers.push(user);
            }
        },
        goToGroupDetails() {
            if (this.selectedUsers.length < 2) return;
            this.viewMode = 'group_details';
            this.groupName = "";
        },
        goBack() {
             this.errorMessage = "";
             if (this.viewMode === 'new_contact' || this.viewMode === 'new_group') this.viewMode = 'new_chat';
             else if (this.viewMode === 'group_details') this.viewMode = 'new_group';
             else this.viewMode = 'list';
        },
        async createDM() {
            if (!this.newContactName) return;
            this.errorMessage = "";
            try {
                let res = await this.$axios.post("/conversations", { recipientName: this.newContactName });
                this.newContactName = "";
                this.viewMode = 'list';
                this.$emit('chat-created', res.data.conversationId);
            } catch (e) {
                if (e.response && e.response.status === 404) {
                    this.errorMessage = "User not found";
                } else if (e.response && e.response.status === 400) {
                     this.errorMessage = "Invalid user (cannot chat with yourself)";
                } else {
                    this.errorMessage = "Error: " + (e.response ? e.response.statusText : e.message);
                }
            }
        },
        async finalizeGroup() {
            if (!this.groupName) return;
            try {
                const memberIds = this.selectedUsers.map(u => u.id);
                let res = await this.$axios.post("/groups", { 
                    name: this.groupName, 
                    initialMembers: memberIds 
                });
                // Group created
                this.viewMode = 'list';
                this.selectedUsers = [];
                this.groupName = "";
                this.$emit('chat-created', res.data.groupId);
            } catch (e) {
                this.$emit('show-error', "Error creating group: " + e.toString());
            }
        },
        resolvePhotoUrl(url) {
            if (!url) return null;
            if (url.startsWith("http")) return url;
            if (url.startsWith("/")) {
                const base = this.$axios.defaults.baseURL || "";
                return base + url;
            }
            return url;
        },
        formatTime(timeStr) {
            if (!timeStr) return "";
            const date = new Date(timeStr);
            return date.toLocaleTimeString([], { hour: 'numeric', minute: '2-digit', hour12: true }).toLowerCase();
        },
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
                            <small class="text-secondary" style="font-size: 0.75rem;">{{ formatTime(c.latestMessageTime) }}</small>
                        </div>
                        <div class="d-flex justify-content-between align-items-center">
                            <div class="d-flex align-items-center gap-1 overflow-hidden" style="max-width: 90%;">
                                <div v-if="parseInt(c.latestMessageSenderId) === parseInt(currentUserId)" class="d-flex align-items-center flex-shrink-0">
                                     <!-- Sent (Clock - Image 2) -->
                                     <svg v-if="c.latestMessageStatus === 0" xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="#8696a0" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>
                                     <!-- Received (Single Check - Image 1) -->
                                     <svg v-else-if="c.latestMessageStatus === 1" xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="#8696a0" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>
                                     <!-- Read (Double Check - Image 3) -->
                                     <svg v-else xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#53bdeb" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="7 12 12 17 22 7"></polyline><polyline points="2 12 7 17 17 7"></polyline></svg>
                                </div>
                                <small v-if="c.latestMessageDeleted" class="text-secondary text-truncate d-block fst-italic">
                                     <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-slash me-1"><circle cx="12" cy="12" r="10"></circle><line x1="4.93" y1="4.93" x2="19.07" y2="19.07"></line></svg>
                                     This Message has been deleted
                                </small>
                                <small v-else class="text-secondary text-truncate d-block">{{ c.latestMessagePreview }}</small>
                            </div>
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
                <input type="text" class="form-control form-control-sm bg-dark-input text-white border-0 mb-3" placeholder="Search name or number">
                
                <div class="d-flex align-items-center p-3 text-white chat-item rounded" @click="openNewGroup">
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

        <!-- VIEW: NEW GROUP (Member Selection) -->
        <div v-else-if="viewMode === 'new_group'" class="d-flex flex-column h-100 position-relative">
            <div class="p-3 bg-dark-header d-flex align-items-center gap-3">
                 <button class="btn btn-link text-white p-0" @click="goBack">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-arrow-left"><line x1="19" y1="12" x2="5" y2="12"></line><polyline points="12 19 5 12 12 5"></polyline></svg>
                 </button>
                <h5 class="m-0 text-white">Add group members</h5>
            </div>

            <div class="p-3 border-bottom border-secondary d-flex flex-wrap gap-2 sticky-top bg-dark-list" v-if="selectedUsers.length > 0">
                <div v-for="user in selectedUsers" :key="user.id" class="badge rounded-pill bg-dark d-flex align-items-center gap-2 p-2 border border-secondary">
                    <div class="rounded-circle bg-secondary overflow-hidden" style="width: 24px; height: 24px;">
                        <img v-if="user.photoUrl" :src="resolvePhotoUrl(user.photoUrl)" class="w-100 h-100" style="object-fit: cover;">
                        <span v-else class="small">{{ user.name.charAt(0) }}</span>
                    </div>
                    <span>{{ user.name }}</span>
                    <button class="btn btn-sm btn-link text-white p-0 lh-1" @click="toggleUserSelection(user)">×</button>
                </div>
            </div>

            <div class="p-2 border-bottom border-secondary">
                <input type="text" class="form-control form-control-sm bg-dark-input text-white border-0" placeholder="Type contact name" v-model="groupSearch">
            </div>

            <div class="flex-grow-1 overflow-auto custom-scrollbar">
                <div v-for="user in filteredUsers" :key="user.id" class="d-flex align-items-center p-3 border-bottom border-secondary chat-item" @click="toggleUserSelection(user)">
                    <div class="rounded-circle bg-secondary d-flex justify-content-center align-items-center me-3 text-white" style="width: 45px; height: 45px; flex-shrink: 0;">
                        <img v-if="user.photoUrl" :src="resolvePhotoUrl(user.photoUrl)" class="w-100 h-100 rounded-circle" style="object-fit: cover;">
                        <span v-else>{{ user.name.charAt(0).toUpperCase() }}</span>
                    </div>
                    <div class="flex-grow-1 text-white">{{ user.name }}</div>
                    <div v-if="selectedUsers.find(u => u.id === user.id)" class="text-success">
                        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-check-circle"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>
                    </div>
                </div>
            </div>

            <!-- Floating Action Button -->
            <button 
                v-if="selectedUsers.length >= 2" 
                class="btn rounded-circle p-3 position-absolute bottom-0 end-0 m-4 shadow-lg d-flex align-items-center justify-content-center" 
                style="background-color: #00a884; width: 60px; height: 60px; z-index: 10;"
                @click="goToGroupDetails"
            >
                <svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#fff" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-arrow-right"><line x1="5" y1="12" x2="19" y2="12"></line><polyline points="12 5 19 12 12 19"></polyline></svg>
            </button>
        </div>

        <!-- VIEW: GROUP DETAILS -->
        <div v-else-if="viewMode === 'group_details'" class="d-flex flex-column h-100">
            <div class="p-3 bg-dark-header d-flex align-items-center gap-3">
                 <button class="btn btn-link text-white p-0" @click="goBack">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-arrow-left"><line x1="19" y1="12" x2="5" y2="12"></line><polyline points="12 19 5 12 12 5"></polyline></svg>
                 </button>
                <h5 class="m-0 text-white">New group</h5>
            </div>

            <div class="p-4 d-flex flex-column align-items-center">
                 <!-- Group Photo Placeholder -->
                 <div class="rounded-circle bg-secondary d-flex justify-content-center align-items-center mb-4 text-white-50 shadow-sm" style="width: 150px; height: 150px; background-color: #202c33 !important;">
                      <svg xmlns="http://www.w3.org/2000/svg" width="60" height="60" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="feather feather-camera"><path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V7a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z"></path><circle cx="12" cy="13" r="4"></circle></svg>
                 </div>

                 <div class="w-100 mb-4">
                     <label class="form-label text-secondary small">Group name</label>
                     <input 
                        v-model="groupName" 
                        type="text" 
                        class="form-control bg-dark-input text-white border-0 border-bottom border-success rounded-0 px-0 py-2" 
                        placeholder="Group name"
                        maxlength="20"
                     >
                 </div>

                 <div class="w-100 text-secondary small mb-4">
                     Provide a group subject and optional group icon
                 </div>

                 <button 
                    class="btn btn-success rounded-circle p-3 mt-auto shadow-lg" 
                    style="width: 60px; height: 60px; background-color: #00a884 !important;" 
                    @click="finalizeGroup"
                    :disabled="!groupName || groupName.length < 3"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#fff" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-check"><polyline points="20 6 9 17 4 12"></polyline></svg>
                 </button>
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
