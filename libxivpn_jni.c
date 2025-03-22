#include <jni.h>
#include <stdlib.h>
#include <android/log.h>
#include <unistd.h>
#include <pthread.h>
#include <netinet/in.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>

// https://github.com/golang/mobile/blob/master/bind/java/seq_android.c.support
// https://developer.android.com/training/articles/perf-jni#threads

static JavaVM *jvm;
static pthread_key_t jnienvs;
static jobject socketProtect = NULL;
static jmethodID socketProtectMethod;

void libxivpn_log(char * msg) {
	__android_log_write(ANDROID_LOG_INFO, "libxivpn", msg);
	free(msg);
}

static void env_destructor(void *env) {
	__android_log_print(ANDROID_LOG_DEBUG, "libxivpn", "detech current thread: %d", gettid());
	if ((*jvm)->DetachCurrentThread(jvm) != JNI_OK) {
		__android_log_write(ANDROID_LOG_ERROR, "libxivpn", "failed to detach current thread");
	}
}

JNIEnv* get_jni_env() {
	JNIEnv *env;
	jint ret = (*jvm)->GetEnv(jvm, (void **)&env, JNI_VERSION_1_6);
	if (ret != JNI_OK) {
		if (ret != JNI_EDETACHED) {
			__android_log_write(ANDROID_LOG_FATAL, "libxivpn", "failed to get thread env");
		}
		__android_log_print(ANDROID_LOG_DEBUG, "libxivpn", "attach current thread: %d", gettid());
		if ((*jvm)->AttachCurrentThread(jvm, &env, NULL) != JNI_OK) {
			__android_log_write(ANDROID_LOG_FATAL, "libxivpn", "failed to attach current thread");
		}
		if (pthread_key_create(&jnienvs, env_destructor) != 0) {
			__android_log_write(ANDROID_LOG_FATAL,  "libxivpn", "failed to initialize jnienvs thread local storage");
		}
		pthread_setspecific(jnienvs, env);
	}
	return env;
}

// Call VpnService.protect
void libxivpn_protect(int fd) {

	__android_log_print(ANDROID_LOG_DEBUG, "libxivpn", "protect %d", fd);

	if (socketProtect == NULL) {
		__android_log_write(ANDROID_LOG_ERROR, "libxivpn", "null socket protect");
		return;
	}

	JNIEnv* env = get_jni_env();


	(*env)->CallVoidMethod(env, socketProtect, socketProtectMethod, fd);
}

extern char* libxivpn_version();
extern char* libxivpn_start(char* config, int socksPort, int fd, char* logFile, char* asset);
extern void libxivpn_stop();


JNIEXPORT jstring JNICALL Java_cn_gov_xivpn2_LibXivpn_xivpn_1version (JNIEnv * env, jclass clazz)
{
	char *v = libxivpn_version();
	jstring s = (*env)->NewStringUTF(env, v);
	free(v);
	return s;
}

JNIEXPORT jstring JNICALL Java_cn_gov_xivpn2_LibXivpn_xivpn_1start (JNIEnv * env, jclass clazz, jstring config, jint socksPort, jint fd, jstring logFile, jstring asset, jobject _socketProtect) {
	char *cConfig = (*env)->GetStringUTFChars(env, config, 0);
	char *cLogFile = (*env)->GetStringUTFChars(env, logFile, 0);
	char *cAsset = (*env)->GetStringUTFChars(env, asset, 0);

	if (socketProtect != NULL) {
		(*env)->DeleteGlobalRef(env, socketProtect);
		socketProtect = NULL;
	}
	socketProtect = (*env)->NewGlobalRef(env, _socketProtect);

	char *ret = libxivpn_start(cConfig, (int)socksPort, (int)fd, cLogFile, cAsset);

	(*env)->ReleaseStringUTFChars(env, config, cConfig);
	(*env)->ReleaseStringUTFChars(env, config, cLogFile);
	(*env)->ReleaseStringUTFChars(env, config, cAsset);

	jstring ret_jstring = (*env)->NewStringUTF(env, ret);
	free(ret);

	return ret_jstring;
}

JNIEXPORT void JNICALL Java_cn_gov_xivpn2_LibXivpn_xivpn_1stop (JNIEnv * env, jclass clazz) {
	libxivpn_stop();
}

JNIEXPORT void JNICALL Java_cn_gov_xivpn2_LibXivpn_xivpn_1init (JNIEnv * env , jclass clazz) {
	__android_log_write(ANDROID_LOG_DEBUG, "libxivpn", "init");

	if ((*env)->GetJavaVM(env, &jvm) != 0) {
		__android_log_write(ANDROID_LOG_FATAL, "libxivpn", "failed to get jvm");
	}

	jclass socketProtectClass = (*env)->FindClass(env, "cn/gov/xivpn2/service/XiVPNService");
	if (clazz == NULL) {
		__android_log_write(ANDROID_LOG_ERROR, "libxivpn", "init: failed to find class");
		return;
	}

	socketProtectMethod = (*env)->GetMethodID(env, socketProtectClass, "protectFd", "(I)V");
	if (clazz == NULL) {
		__android_log_write(ANDROID_LOG_ERROR, "libxivpn", "init: failed to get method id");
		return;
	}

}

jint JNI_OnLoad(JavaVM *vm, void *reserved) {
	__android_log_write(ANDROID_LOG_DEBUG, "libxivpn", "jni onload");
	return JNI_VERSION_1_6;
}