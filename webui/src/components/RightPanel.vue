<script>
export default {
    props: ['conversation', 'username', 'userId'],
    emits: ['set-group-name', 'set-group-photo', 'add-member', 'leave-group', 'close'],
    data() {
        return {
            view: 'info', // 'info', 'add_member'
            isEditingName: false,
            editNameValue: this.conversation.name || "",
            members: [],
            allUsers: [],
            selectedUsers: [],
            memberSearch: ""
        }
    },
    watch: {
        conversation: {
            handler(val) {
                this.editNameValue = val.name || "";
                if (val.isGroup) this.fetchMembers();
            },
            deep: true,
            immediate: true
        }
    },
    computed: {
        filteredAvailableUsers() {
            const query = this.memberSearch.toLowerCase();
            return this.allUsers.filter(u => 
                u.name.toLowerCase().includes(query) && 
                u.id.toString() !== this.userId?.toString()
            );
        }
    },
    methods: {
        saveName() {
            if (this.editNameValue !== this.conversation.name) {
                this.$emit('set-group-name', this.conversation.conversationId, this.editNameValue);
            }
            this.isEditingName = false;
        },
        openPhotoUpload() {
             this.$refs.fileInput.click();
        },
        handleFileChange(event) {
            const file = event.target.files[0];
            if (file) {
                 this.$emit('set-group-photo', this.conversation.conversationId, file);
            }
        },
        async fetchMembers() {
            try {
                let res = await this.$axios.get(`/groups/${this.conversation.conversationId}/members`);
                this.members = res.data;
            } catch (e) {
                console.error("Error fetching members", e);
            }
        },
        async openAddMember() {
            this.view = 'add_member';
            this.selectedUsers = [];
            this.memberSearch = "";
            try {
                let res = await this.$axios.get("/users");
                this.allUsers = res.data;
            } catch (e) {
                console.error("Error fetching all users", e);
            }
        },
        toggleUserSelection(user) {
            if (this.isAlreadyInGroup(user.id)) return;
            const idx = this.selectedUsers.findIndex(u => u.id === user.id);
            if (idx > -1) {
                this.selectedUsers.splice(idx, 1);
            } else {
                this.selectedUsers.push(user);
            }
        },
        isAlreadyInGroup(userId) {
            return this.members.some(m => m.id === userId);
        },
        async finalizeAddMembers() {
            if (this.selectedUsers.length === 0) return;
            try {
                const ids = this.selectedUsers.map(u => u.id);
                await this.$axios.post(`/groups/${this.conversation.conversationId}/members`, { userIds: ids });
                this.view = 'info';
                this.fetchMembers();
            } catch (e) {
                alert("Error adding members: " + e.toString());
            }
        },
        leaveGroup() {
            if (confirm("Are you sure you want to leave this group?")) {
                this.$emit('leave-group', this.conversation.conversationId);
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
    }
}
</script>

<template>
    <div class="d-flex flex-column h-100 bg-dark-list border-start border-secondary" style="width: 300px;">
        <input type="file" ref="fileInput" class="d-none" accept="image/*" @change="handleFileChange">
        <div v-if="view === 'info'" class="flex-grow-1 d-flex flex-column overflow-hidden">
            <!-- Header -->
            <div class="p-3 bg-dark-header d-flex align-items-center">
                <button class="btn btn-link text-secondary p-0 me-3" @click="$emit('close')">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-x"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                </button>
                <h5 class="m-0 text-white">{{ conversation.isGroup ? 'Group Info' : 'Contact Info' }}</h5>
            </div>

            <div class="flex-grow-1 p-0 overflow-auto custom-scrollbar bg-dark-list">
                <!-- Photo/Name Section -->
                <div class="p-3 d-flex flex-column align-items-center mb-2 bg-dark-header py-4">
                    <div class="position-relative mb-3" style="width: 200px; height: 200px; cursor: pointer;" @click="conversation.isGroup ? openPhotoUpload() : null" :title="conversation.isGroup ? 'Change Group Photo' : ''">
                        <div class="w-100 h-100 rounded-circle bg-secondary d-flex align-items-center justify-content-center overflow-hidden">
                            <img v-if="conversation.photoUrl" :src="resolvePhotoUrl(conversation.photoUrl)" class="w-100 h-100" style="object-fit: cover;">
                            <span v-else class="text-white" style="font-size: 5rem;">{{ (conversation.name || 'C').charAt(0).toUpperCase() }}</span>
                        </div>
                        <div v-if="conversation.isGroup" class="position-absolute top-0 start-0 w-100 h-100 rounded-circle d-flex align-items-center justify-content-center bg-dark bg-opacity-50 text-white opacity-0 hover-opacity-100 transition">
                            <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-camera"><path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z"></path><circle cx="12" cy="13" r="4"></circle></svg>
                        </div>
                    </div>
                    
                    <div class="w-100 text-center">
                        <div v-if="!isEditingName" class="d-flex justify-content-center align-items-center gap-2 mb-1">
                            <h4 class="text-white m-0 text-break">{{ conversation.name || (conversation.isGroup ? 'Group ' + conversation.conversationId : 'User ' + conversation.conversationId) }}</h4>
                            <button v-if="conversation.isGroup" class="btn btn-link text-secondary p-0" @click="isEditingName = true">
                                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-edit-2"><path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"></path></svg>
                            </button>
                        </div>
                        <div v-else class="d-flex gap-2 justify-content-center mb-1">
                            <input v-model="editNameValue" class="form-control form-control-sm bg-dark-input text-white border-bottom border-success border-0 rounded-0 p-0 text-center" @keyup.enter="saveName">
                            <button class="btn btn-link text-success p-0" @click="saveName">
                                <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-check"><polyline points="20 6 9 17 4 12"></polyline></svg>
                            </button>
                        </div>
                        <small class="text-secondary" v-if="conversation.isGroup">Group • {{ members.length }} participants</small>
                    </div>
                </div>

                <!-- Group Members Section -->
                <div v-if="conversation.isGroup" class="bg-dark-header mt-2 mb-2 p-3">
                    <div class="d-flex align-items-center gap-3 p-2 chat-item rounded mb-2" @click="openAddMember">
                        <div class="rounded-circle bg-success d-flex justify-content-center align-items-center" style="width: 40px; height: 40px;">
                            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-user-plus"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="8.5" cy="7" r="4"></circle><line x1="20" y1="8" x2="20" y2="14"></line><line x1="23" y1="11" x2="17" y2="11"></line></svg>
                        </div>
                        <span class="text-white">Add member</span>
                    </div>

                    <div v-for="m in members" :key="m.id" class="d-flex align-items-center gap-3 p-2 py-3 border-top border-secondary">
                        <div class="rounded-circle bg-secondary overflow-hidden" style="width: 40px; height: 40px; flex-shrink: 0;">
                            <img v-if="m.photoUrl" :src="resolvePhotoUrl(m.photoUrl)" class="w-100 h-100" style="object-fit: cover;">
                            <span v-else class="w-100 h-100 d-flex align-items-center justify-content-center text-white">{{ m.name.charAt(0).toUpperCase() }}</span>
                        </div>
                        <div class="flex-grow-1 min-w-0">
                            <div class="text-white text-truncate">{{ parseInt(m.id) === parseInt(userId) ? 'You' : m.name }}</div>
                            <small class="text-secondary d-block text-truncate">Hey there! I am using WASAText.</small>
                        </div>
                    </div>
                </div>

                <!-- Exit Group -->
                <div v-if="conversation.isGroup" class="bg-dark-header p-3 mt-2">
                    <button class="btn btn-link text-danger w-100 d-flex align-items-center gap-3 p-0 text-decoration-none" @click="leaveGroup">
                        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-log-out"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path><polyline points="16 17 21 12 16 7"></polyline><line x1="21" y1="12" x2="9" y2="12"></line></svg>
                        <span>Exit Group</span>
                    </button>
                </div>
            </div>
        </div>

        <!-- VIEW: ADD MEMBER -->
        <div v-else-if="view === 'add_member'" class="flex-grow-1 d-flex flex-column bg-dark-list">
             <!-- Header -->
             <div class="p-3 bg-dark-header d-flex align-items-center gap-3">
                 <button class="btn btn-link text-white p-0" @click="view = 'info'">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-arrow-left"><line x1="19" y1="12" x2="5" y2="12"></line><polyline points="12 19 5 12 12 5"></polyline></svg>
                 </button>
                <h5 class="m-0 text-white">Add members</h5>
            </div>

            <div class="p-2 border-bottom border-secondary">
                <input type="text" class="form-control form-control-sm bg-dark-input text-white border-0" placeholder="Type contact name" v-model="memberSearch">
            </div>

            <div class="flex-grow-1 overflow-auto custom-scrollbar">
                <div v-for="user in filteredAvailableUsers" :key="user.id" 
                    class="d-flex align-items-center p-3 border-bottom border-secondary chat-item" 
                    :class="{ 'opacity-50': isAlreadyInGroup(user.id) }"
                    @click="toggleUserSelection(user)"
                >
                    <div class="rounded-circle bg-secondary d-flex justify-content-center align-items-center me-3 text-white" style="width: 45px; height: 45px; flex-shrink: 0;">
                        <img v-if="user.photoUrl" :src="resolvePhotoUrl(user.photoUrl)" class="w-100 h-100 rounded-circle" style="object-fit: cover;">
                        <span v-else>{{ user.name.charAt(0).toUpperCase() }}</span>
                    </div>
                    <div class="flex-grow-1 text-white">
                        <div>{{ user.name }}</div>
                        <small v-if="isAlreadyInGroup(user.id)" class="text-secondary">Already in group</small>
                    </div>
                    <div v-if="selectedUsers.find(u => u.id === user.id)" class="text-success">
                        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-check-circle"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>
                    </div>
                </div>
            </div>

            <!-- FAB -->
            <button 
                v-if="selectedUsers.length > 0" 
                class="btn rounded-circle p-3 position-absolute bottom-0 end-0 m-4 shadow-lg d-flex align-items-center justify-content-center btn-success" 
                style="width: 60px; height: 60px; z-index: 10;"
                @click="finalizeAddMembers"
            >
                <svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#fff" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-check"><polyline points="20 6 9 17 4 12"></polyline></svg>
            </button>
        </div>
    </div>
</template>

<style scoped>
.bg-dark-list { background-color: #0c1317; }
.bg-dark-header { background-color: #111b21; }
.bg-dark-input { background-color: #2a3942; }

.chat-item { cursor: pointer; transition: background-color 0.2s; }
.chat-item:hover { background-color: #202c33; }

.opacity-50 { opacity: 0.5; cursor: not-allowed !important; }

.text-secondary { color: #8696a0 !important; }
.text-success { color: #00a884 !important; }
.btn-success { background-color: #00a884 !important; border: none; }

.hover-opacity-100:hover { opacity: 1 !important; }
.transition { transition: opacity 0.2s; }
</style>
